package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/crypto/bls"
	"github.com/zarbchain/zarb-go/crypto/hash"
	"github.com/zarbchain/zarb-go/tx"
)

func TestExecuteUnbondTx(t *testing.T) {
	setup(t)
	exe := NewUnbondExecutor(true)

	addr := crypto.GenerateTestAddress()

	hash100 := hash.GenerateTestHash()
	tSandbox.AppendNewBlock(100, hash100)

	t.Run("Should fail, Invalid validator", func(t *testing.T) {
		trx := tx.NewUnbondTx(hash100.Stamp(), 1, addr, "invalid validator")
		assert.Error(t, exe.Execute(trx, tSandbox))
	})

	t.Run("Should fail, Invalid sequence", func(t *testing.T) {
		trx := tx.NewUnbondTx(hash100.Stamp(), tSandbox.Validator(tVal1.Address()).Sequence()+2, tVal1.Address(), "invalid sequence")

		assert.Error(t, exe.Execute(trx, tSandbox))
	})

	t.Run("Should fail, Inside committee", func(t *testing.T) {
		tSandbox.InCommittee = true
		trx := tx.NewUnbondTx(hash100.Stamp(), tSandbox.Validator(tVal1.Address()).Sequence()+1, tVal1.Address(), "inside committee")

		assert.Error(t, exe.Execute(trx, tSandbox))
	})

	t.Run("Ok", func(t *testing.T) {
		tSandbox.InCommittee = false
		trx := tx.NewUnbondTx(hash100.Stamp(), tSandbox.Validator(tVal1.Address()).Sequence()+1, tVal1.Address(), "Ok")

		assert.NoError(t, exe.Execute(trx, tSandbox))

		// Replay
		assert.Error(t, exe.Execute(trx, tSandbox))
	})

	t.Run("Should fail, Cannot unbond if unbonded already", func(t *testing.T) {
		tSandbox.InCommittee = false
		trx := tx.NewUnbondTx(hash100.Stamp(), tSandbox.Validator(tVal1.Address()).Sequence()+1, tVal1.Address(), "Ok")

		assert.Error(t, exe.Execute(trx, tSandbox))
	})
	assert.Equal(t, tSandbox.Validator(tVal1.Address()).Stake(), int64(5000000000))
	assert.Equal(t, tSandbox.Validator(tVal1.Address()).Power(), int64(0))
	assert.Equal(t, tSandbox.Validator(tVal1.Address()).UnbondingHeight(), 101)
	assert.Equal(t, exe.Fee(), int64(0))

	checkTotalCoin(t, 0)
}

func TestUnbondNonStrictMode(t *testing.T) {
	setup(t)
	exe1 := NewBondExecutor(false)

	tSandbox.InCommittee = true
	hash100 := hash.GenerateTestHash()
	tSandbox.AppendNewBlock(100, hash100)
	bonder := tAcc1.Address()
	pub, _ := bls.GenerateTestKeyPair()

	mintbase1 := tx.NewBondTx(hash100.Stamp(), tSandbox.AccSeq(bonder)+1, bonder, pub, 1000, 1000, "")
	mintbase2 := tx.NewBondTx(hash100.Stamp(), tSandbox.AccSeq(bonder)+1, bonder, pub, 1000, 1000, "")

	assert.NoError(t, exe1.Execute(mintbase1, tSandbox))
	assert.Error(t, exe1.Execute(mintbase2, tSandbox)) // Invalid sequence
}