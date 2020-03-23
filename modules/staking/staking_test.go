package staking_test

import (
	"testing"

	sdk "github.com/irisnet/irishub-sdk-go/types"

	"github.com/stretchr/testify/suite"

	"github.com/irisnet/irishub-sdk-go/test"
)

type StakingTestSuite struct {
	suite.Suite
	test.MockClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(StakingTestSuite))
}

func (sts *StakingTestSuite) SetupTest() {
	sts.MockClient = test.NewMockClient()
}

func (sts *StakingTestSuite) TestStaking() {
	baseTx := sdk.BaseTx{
		From:     sts.Account().Name,
		Gas:      20000,
		Memo:     "test",
		Mode:     sdk.Commit,
		Password: sts.Account().Password,
	}

	//test QueryValidators
	validators, _ := sts.Staking().QueryValidators(1, 10)
	validator := validators[0].OperatorAddress

	amount, e := sdk.ParseDecCoin("20iris")
	sts.NoError(e)

	//test Delegate
	rs, err := sts.Staking().Delegate(validator, amount, baseTx)
	sts.NoError(err)
	sts.NotEmpty(rs.Hash)

	//test QueryDelegation
	delegator := sts.Account().Address.String()
	d, err := sts.Staking().QueryDelegation(delegator, validator)
	sts.NoError(err)
	sts.Equal(validator, d.ValidatorAddr)
	sts.Equal(delegator, d.DelegatorAddr)

	//test QueryDelegations
	ds, err := sts.Staking().QueryDelegations(delegator)
	sts.NoError(err)
	sts.NotEmpty(ds)

	//test QueryDelegationsTo
	ds, err = sts.Staking().QueryDelegationsTo(validator)
	sts.NoError(err)
	sts.NotEmpty(ds)

	//test Undelegate
	amount, e = sdk.ParseDecCoin("10iris")
	sts.NoError(e)

	rs, err = sts.Staking().Undelegate(validator, amount, baseTx)
	sts.NoError(err)
	sts.NotEmpty(rs.Hash)

	//test QueryUnbondingDelegation
	ubd, err := sts.Staking().QueryUnbondingDelegation(delegator, validator)
	sts.NoError(err)
	sts.Equal(validator, ubd.ValidatorAddr)
	sts.Equal(delegator, ubd.DelegatorAddr)

	//test QueryUnbondingDelegations
	ubds, err := sts.Staking().QueryUnbondingDelegations(delegator)
	sts.NoError(err)
	sts.NotEmpty(ubds)

	//test QueryUnbondingDelegationsFrom
	uds, err := sts.Staking().QueryUnbondingDelegationsFrom(validator)
	sts.NoError(err)
	sts.NotEmpty(uds)
}

//func (sts *StakingTestSuite) TestDelegate() {
//	baseTx := sdk.BaseTx{
//		From: "test1",
//		Gas:  20000,
//		Memo: "test",
//		Mode: sdk.Commit,
//	}
//
//	validators, _ := sts.Staking().QueryValidators(1, 10)
//	validator := validators[0].OperatorAddress
//	amt, _ := sdk.NewIntFromString("20000000000000000000")
//	amount := sdk.NewCoin("iris-atto", amt)
//
//	rs, err := sts.Staking().Delegate(validator, amount, baseTx)
//	require.NoError(err)
//	require.NotEmpty(rs.Hash)
//
//}

//func (sts *StakingTestSuite) TestUndelegate() {
//	baseTx := sdk.BaseTx{
//		From: "test1",
//		Gas:  20000,
//		Memo: "test",
//		Mode: sdk.Commit,
//	}
//
//	validators, _ := sts.Staking().QueryValidators(1, 10)
//	validator := validators[0].OperatorAddress
//	amt, _ := sdk.NewIntFromString("10000000000000000000")
//	amount := sdk.NewCoin("iris-atto", amt)
//
//	rs, err := sts.Staking().Undelegate(validator, amount, baseTx)
//	require.NoError(err)
//	require.NotEmpty(rs.Hash)
//
//
//}

//
//func (sts *StakingTestSuite) TestRedelegate() {
//	baseTx := sdk.BaseTx{
//		From: "test1",
//		Gas:  20000,
//		Memo: "test",
//		Mode: sdk.Commit,
//	}
//
//	validators, _ := sts.Staking().QueryValidators(1, 10)
//	validator := validators[0].OperatorAddress
//	amt, _ := sdk.NewIntFromString("1000000000000000000")
//	amount := sdk.NewCoin("iris-atto", amt)
//
//	rs, err := sts.Staking().Undelegate(validator, amount, baseTx)
//	require.NoError(err)
//	require.NotEmpty(rs.Hash)
//
//}

//func (sts *StakingTestSuite) TestQueryDelegation() {
//	validators, _ := sts.Staking().QueryValidators(1, 10)
//	validator := validators[0].OperatorAddress
//	delegator := sts.Sender().String()
//	d, err := sts.Staking().QueryDelegation(delegator, validator)
//	require.NoError(err)
//	require.EQ(validator, d.ValidatorAddr)
//	require.EQ(delegator, d.DelegatorAddr)
//}
//
//func (sts *StakingTestSuite) TestQueryQueryDelegations() {
//	delegator := sts.Sender().String()
//	ds, err := sts.Staking().QueryDelegations(delegator)
//	require.NoError(err)
//	require.NotEmpty(ds)
//}
//
//func (sts *StakingTestSuite) TestQueryUnbondingDelegation() {
//	validators, _ := sts.Staking().QueryValidators(1, 10)
//	validator := validators[0].OperatorAddress
//	delegator := sts.Sender().String()
//	ubd, err := sts.Staking().QueryUnbondingDelegation(delegator, validator)
//	require.NoError(err)
//	require.EQ(validator, ubd.ValidatorAddr)
//	require.EQ(delegator, ubd.DelegatorAddr)
//}
//
//func (sts *StakingTestSuite) TestQueryUnbondingDelegations() {
//	delegator := sts.Sender().String()
//	ubds, err := sts.Staking().QueryUnbondingDelegations(delegator)
//	require.NoError(err)
//	require.NotEmpty(ubds)
//}
//
//func (sts *StakingTestSuite) TestQueryDelegationsTo() {
//	validators, _ := sts.Staking().QueryValidators(1, 10)
//	validator := validators[0].OperatorAddress
//	ds, err := sts.Staking().QueryDelegationsTo(validator)
//	require.NoError(err)
//	require.NotEmpty(ds)
//}
//
//func (sts *StakingTestSuite) TestQueryUnbondingDelegationsFrom() {
//	validators, _ := sts.Staking().QueryValidators(1, 10)
//	validator := validators[0].OperatorAddress
//	ds, err := sts.Staking().QueryUnbondingDelegationsFrom(validator)
//	require.NoError(err)
//	require.NotEmpty(ds)
//}
//
//func (sts *StakingTestSuite) TestQueryValidator() {
//	validators, _ := sts.Staking().QueryValidators(1, 10)
//	validator := validators[0].OperatorAddress
//	v, err := sts.Staking().QueryValidator(validator)
//	require.NoError(err)
//	require.EqualValues(validators[0], v)
//}

func (sts *StakingTestSuite) TestQueryPool() {
	p, err := sts.Staking().QueryPool()
	sts.NoError(err)
	sts.NotEmpty(p)
}

func (sts *StakingTestSuite) TestQueryParams() {
	p, err := sts.Staking().QueryParams()
	sts.NoError(err)
	sts.NotEmpty(p)
}
