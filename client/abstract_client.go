package client

import (
	"errors"
	"fmt"

	"github.com/irisnet/irishub-sdk-go/tools/log"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	sdk "github.com/irisnet/irishub-sdk-go/types"
)

type abstractClient struct {
	*sdk.TxContext
	sdk.RPC
	logger *log.Logger
}

func (ac abstractClient) Logger() *log.Logger {
	return ac.logger
}

func (ac abstractClient) BuildAndSend(msg []sdk.Msg, baseTx sdk.BaseTx) (sdk.Result, error) {
	//validate msg
	for _, m := range msg {
		if err := m.ValidateBasic(); err != nil {
			return nil, err
		}
	}
	ac.Logger().Info().Msg("validate msg success")

	err := ac.prepareTxContext(baseTx)
	if err != nil {
		return nil, err
	}

	tx, err := ac.BuildAndSign(baseTx.From, msg)
	if err != nil {
		return nil, err
	}
	ac.Logger().Info().Msg("sign transaction success")

	txByte, err := ac.Codec.MarshalBinaryLengthPrefixed(tx)
	if err != nil {
		return nil, err
	}

	rs, err := ac.broadcastTx(txByte)
	if err == nil {
		ac.Logger().Info().
			Str("mode", string(ac.Mode)).
			Str("hash", rs.GetHash()).
			Msg("broadcast transaction success")
	}
	return rs, err
}

func (ac abstractClient) Broadcast(signedTx sdk.StdTx, mode sdk.BroadcastMode) (sdk.Result, error) {
	ac.Mode = mode
	txByte, err := ac.Codec.MarshalBinaryLengthPrefixed(signedTx)
	if err != nil {
		return nil, err
	}

	return ac.broadcastTx(txByte)
}

func (ac abstractClient) Sign(stdTx sdk.StdTx, name string, password string, online bool) (sdk.StdTx, error) {
	baseTx := sdk.BaseTx{
		From:     name,
		Password: password,
		Gas:      stdTx.Fee.Gas,
		Fee:      stdTx.Fee.Amount.String(),
		Memo:     stdTx.Memo,
	}
	ac.WithOnline(online)
	err := ac.prepareTxContext(baseTx)
	if err != nil {
		return stdTx, err
	}

	tx, err := ac.BuildAndSign(baseTx.From, stdTx.GetMsgs())
	if err != nil {
		return stdTx, err
	}

	return tx, nil
}

func (ac abstractClient) QueryWithResponse(path string, data interface{}, result sdk.Response) error {
	var bz []byte
	var err error
	if data != nil {
		bz, err = ac.Codec.MarshalJSON(data)
		if err != nil {
			return err
		}
	}

	res, err := ac.RPC.Query(path, bz)
	if err != nil {
		return err
	}

	if err = ac.Codec.UnmarshalJSON(res, result); err != nil {
		return err
	}

	return nil
}

func (ac abstractClient) Query(path string, data interface{}) ([]byte, error) {
	var bz []byte
	var err error
	if data != nil {
		bz, err = ac.Codec.MarshalJSON(data)
		if err != nil {
			return nil, err
		}
	}

	res, err := ac.RPC.Query(path, bz)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ac abstractClient) QueryStore(key cmn.HexBytes, storeName string) (res []byte, err error) {
	path := fmt.Sprintf("/store/%s/%s", storeName, "subspace")
	opts := rpcclient.ABCIQueryOptions{
		//Height: cliCtx.Height,
		Prove: false,
	}

	result, err := ac.RPC.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return res, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return res, errors.New(resp.Log)
	}
	return resp.Value, nil
}

func (ac abstractClient) QueryAccount(address string) (baseAccount sdk.BaseAccount, err error) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return baseAccount, err
	}
	param := struct {
		Address sdk.AccAddress
	}{
		Address: addr,
	}
	if err = ac.QueryWithResponse("custom/acc/account", param, &baseAccount); err != nil {
		return baseAccount, err
	}
	return
}

func (ac abstractClient) QueryAddress(name, password string) (addr sdk.AccAddress, err error) {
	return (*ac.TxContext).KeyManager.QueryAddress(name, password)
}

func (ac abstractClient) prepareTxContext(baseTx sdk.BaseTx) error {
	ctx := ac.TxContext
	if ctx.Online {
		addr, err := ac.QueryAddress(baseTx.From, baseTx.Password)
		if err != nil {
			return err
		}
		account, err := ac.QueryAccount(addr.String())
		ctx = ctx.WithAccountNumber(account.AccountNumber).
			WithSequence(account.Sequence)
	}
	if len(baseTx.Fee) > 0 {
		fee, err := sdk.ParseCoins(baseTx.Fee)
		if err != nil {
			return err
		}
		ctx = ctx.WithFee(fee)
	}

	if len(baseTx.Mode) > 0 {
		ctx = ctx.WithMode(baseTx.Mode)
	}

	if baseTx.Simulate {
		ctx = ctx.WithSimulate(baseTx.Simulate)
	}

	ctx = ctx.WithGas(baseTx.Gas)
	ctx = ctx.WithMemo(baseTx.Memo)
	return nil
}
func (ac abstractClient) broadcastTx(txBytes []byte) (sdk.Result, error) {
	switch ac.Mode {
	case sdk.Commit:
		return ac.broadcastTxCommit(txBytes)
	case sdk.Async:
		return ac.broadcastTxAsync(txBytes)
	case sdk.Sync:
		return ac.broadcastTxSync(txBytes)

	}
	panic("invalid broadcast mode")
}

// broadcastTxCommit broadcasts transaction bytes to a Tendermint node
// and waits for a commit.
func (ac abstractClient) broadcastTxCommit(tx []byte) (result sdk.ResultBroadcastTxCommit, err error) {
	res, err := ac.RPC.BroadcastTxCommit(tx)
	if err != nil {
		return result, err
	}

	if !res.CheckTx.IsOK() {
		return result, errors.New(res.CheckTx.Log)
	}

	if !res.DeliverTx.IsOK() {
		return result, errors.New(res.DeliverTx.Log)
	}
	return sdk.ResultBroadcastTxCommit{
		CheckTx:   res.CheckTx,
		DeliverTx: res.DeliverTx,
		Hash:      res.Hash,
		Height:    res.Height,
	}, err
}

// BroadcastTxSync broadcasts transaction bytes to a Tendermint node
// synchronously.
func (ac abstractClient) broadcastTxSync(tx []byte) (result sdk.ResultBroadcastTxCommit, err error) {
	res, err := ac.RPC.BroadcastTxSync(tx)
	if err != nil {
		return result, err
	}

	return sdk.ResultBroadcastTxCommit{
		Hash: res.Hash,
		CheckTx: abci.ResponseCheckTx{
			Code: res.Code,
			Data: res.Data,
			Log:  res.Log,
		},
	}, nil
}

// BroadcastTxAsync broadcasts transaction bytes to a Tendermint node
// asynchronously.
func (ac abstractClient) broadcastTxAsync(tx []byte) (result sdk.ResultBroadcastTx, err error) {
	res, err := ac.RPC.BroadcastTxAsync(tx)
	if err != nil {
		return result, err
	}

	return sdk.ResultBroadcastTx{
		Code: res.Code,
		Data: res.Data,
		Log:  res.Log,
		Hash: res.Hash,
	}, nil
}
