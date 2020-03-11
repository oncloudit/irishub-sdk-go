// Package slashing provides the ability to access the interface of the IRISHUB slashing module
//
// In Proof-of-Stake blockchain, validators will get block provisions by staking their token.
// But if they failed to keep online, they will be punished by slashing a small portion of their staked tokens.
// The offline validators will be removed from the validator set and put into jail, which means their voting power is zero.
// During the jail period, these nodes are not even validator candidates. Once the jail period ends, they can send [[unjail]] transactions to free themselves and become validator candidates again.
//
// [More Details](https://www.irisnet.org/docs/features/slashing.html)
//
package slashing

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/irisnet/irishub-sdk-go/rpc"

	"github.com/irisnet/irishub-sdk-go/tools/log"
	sdk "github.com/irisnet/irishub-sdk-go/types"
)

type slashingClient struct {
	sdk.AbstractClient
	*log.Logger
}

func Create(ac sdk.AbstractClient) rpc.Slashing {
	return slashingClient{
		AbstractClient: ac,
		Logger:         ac.Logger().With(ModuleName),
	}
}

//QueryParams return parameter for slashing at genesis
func (s slashingClient) QueryParams() (rpc.SlashingParams, sdk.Error) {
	return s.queryParamsV017()
}

//QueryValidatorSigningInfo return the specified validator sign information
func (s slashingClient) QueryValidatorSigningInfo(validatorConPubKey string) (rpc.ValidatorSigningInfo, sdk.Error) {
	pk, err := sdk.GetConsPubKeyBech32(validatorConPubKey)
	if err != nil {
		return rpc.ValidatorSigningInfo{}, sdk.Wrap(err)
	}
	return s.querySigningInfoV017(pk)
}

func (s slashingClient) RegisterCodec(cdc sdk.Codec) {
	registerCodec(cdc)
}

func (s slashingClient) Name() string {
	return ModuleName
}

func (s slashingClient) queryParamsV017() (rpc.SlashingParams, sdk.Error) {
	param := struct {
		Module string
	}{
		Module: s.Name(),
	}

	var params paramsV017
	if err := s.QueryWithResponse("custom/params/module", param, &params); err != nil {
		return rpc.SlashingParams{}, sdk.Wrap(err)
	}
	return params.Convert().(rpc.SlashingParams), sdk.Nil
}

func (s slashingClient) queryParamsV100() (rpc.SlashingParams, error) {
	var params params
	err := s.QueryWithResponse("custom/%s/parameters", s.Name(), &params)
	if err != nil {
		return rpc.SlashingParams{}, err
	}
	return params.Convert().(rpc.SlashingParams), nil
}

func (s slashingClient) querySigningInfoV017(pk crypto.PubKey) (rpc.ValidatorSigningInfo, sdk.Error) {
	key := append([]byte{0x01}, pk.Bytes()...)
	res, err := s.QueryStore(key, s.Name())
	if err != nil {
		return rpc.ValidatorSigningInfo{}, sdk.Wrap(err)
	}

	var signingInfo validatorSigningInfoV017
	err = cdc.UnmarshalBinaryLengthPrefixed(res, &signingInfo)
	if err != nil {
		return rpc.ValidatorSigningInfo{}, sdk.Wrap(err)
	}

	consAddr := sdk.ConsAddress(pk.Address())
	return rpc.ValidatorSigningInfo{
		Address:             consAddr.String(),
		StartHeight:         signingInfo.StartHeight,
		IndexOffset:         signingInfo.IndexOffset,
		JailedUntil:         signingInfo.JailedUntil,
		Tombstoned:          false,
		MissedBlocksCounter: signingInfo.MissedBlocksCounter,
	}, sdk.Nil
}

func (s slashingClient) querySigningInfoV100(pk crypto.PubKey) (rpc.ValidatorSigningInfo, error) {
	consAddr := sdk.ConsAddress(pk.Address())
	param := struct {
		ConsAddress sdk.ConsAddress
	}{
		ConsAddress: consAddr,
	}

	var signingInfo validatorSigningInfo
	err := s.QueryWithResponse("custom/slashing/signingInfo", param, &signingInfo)
	if err != nil {
		return rpc.ValidatorSigningInfo{}, err
	}
	return signingInfo.Convert().(rpc.ValidatorSigningInfo), err
}