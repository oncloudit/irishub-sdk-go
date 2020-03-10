package service

import (
	"github.com/irisnet/irishub-sdk-go/rpc"
	"github.com/irisnet/irishub-sdk-go/tools/log"
	sdk "github.com/irisnet/irishub-sdk-go/types"
)

type serviceClient struct {
	sdk.AbstractClient
	*log.Logger
}

func (s serviceClient) RegisterCodec(cdc sdk.Codec) {
	registerCodec(cdc)
}

func (s serviceClient) Name() string {
	return ModuleName
}

func Create(ac sdk.AbstractClient) rpc.Service {
	return serviceClient{
		AbstractClient: ac,
		Logger:         ac.Logger().With(ModuleName),
	}
}

//DefineService is responsible for creating a new service definition
func (s serviceClient) DefineService(request rpc.ServiceDefinitionRequest, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	author, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}
	msg := MsgDefineService{
		Name:              request.ServiceName,
		Description:       request.Description,
		Tags:              request.Tags,
		Author:            author,
		AuthorDescription: request.AuthorDescription,
		Schemas:           request.Schemas,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

//BindService is responsible for binding a new service definition
func (s serviceClient) BindService(request rpc.ServiceBindingRequest, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}

	msg := MsgBindService{
		ServiceName: request.ServiceName,
		Provider:    provider,
		Deposit:     request.Deposit,
		Pricing:     request.Pricing,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

//UpdateServiceBinding updates the specified service binding
func (s serviceClient) UpdateServiceBinding(request rpc.UpdateServiceBindingRequest, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}
	msg := MsgUpdateServiceBinding{
		ServiceName: request.ServiceName,
		Provider:    provider,
		Deposit:     request.Deposit,
		Pricing:     request.Pricing,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// DisableServiceBinding disables the specified service binding
func (s serviceClient) DisableServiceBinding(serviceName string, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}
	msg := MsgDisableService{
		ServiceName: serviceName,
		Provider:    provider,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// EnableServiceBinding enables the specified service binding
func (s serviceClient) EnableServiceBinding(serviceName string, deposit sdk.Coins, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}
	msg := MsgEnableService{
		ServiceName: serviceName,
		Provider:    provider,
		Deposit:     deposit,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

//InvokeService is responsible for invoke a new service and callback `callback`
func (s serviceClient) InvokeService(request rpc.ServiceInvocationRequest,
	callback rpc.ServiceInvokeHandler, baseTx sdk.BaseTx) (string, sdk.Error) {
	consumer, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return "", sdk.Wrap(err)
	}

	var providers []sdk.AccAddress
	for _, provider := range request.Providers {
		p, err := sdk.AccAddressFromBech32(provider)
		if err != nil {
			return "", sdk.Wrapf("%s invalid address", p)
		}
		providers = append(providers, p)
	}

	msg := MsgRequestService{
		ServiceName:       request.ServiceName,
		Providers:         providers,
		Consumer:          consumer,
		Input:             request.Input,
		ServiceFeeCap:     request.ServiceFeeCap,
		Timeout:           request.Timeout,
		SuperMode:         request.SuperMode,
		Repeated:          request.Repeated,
		RepeatedFrequency: request.RepeatedFrequency,
		RepeatedTotal:     request.RepeatedTotal,
	}

	//mode must be set to commit
	baseTx.Mode = sdk.Commit

	result, err1 := s.BuildAndSend([]sdk.Msg{msg}, baseTx)
	if !err1.IsNil() {
		return "", err1
	}

	requestContextID := result.GetTags().GetValue(tagRequestContextID)
	if callback == nil {
		return requestContextID, sdk.Nil
	}
	builder := sdk.NewEventQueryBuilder().
		AddCondition(sdk.ActionKey, tagRespondService).
		AddCondition(tagConsumer, sdk.EventValue(consumer.String())).
		AddCondition(tagServiceName, sdk.EventValue(request.ServiceName))

	var subscription sdk.Subscription
	subscription, err = s.SubscribeTx(builder, func(tx sdk.EventDataTx) {
		s.Debug().Str("tx_hash", tx.Hash).Int64("height", tx.Height).
			Msg("consumer received response transaction sent by provider")
		for _, msg := range tx.Tx.Msgs {
			msg, ok := msg.(MsgRespondService)
			if ok {
				callback(requestContextID, msg.Output)
			}
		}
		//cancel subscription
		if !request.Repeated {
			_ = s.Unscribe(subscription)
		}
	})
	if err != nil {
		return "", sdk.Wrap(err)
	}

	return requestContextID, sdk.Nil
}

// SetWithdrawAddress sets a new withdrawal address for the specified service binding
func (s serviceClient) SetWithdrawAddress(withdrawAddress string, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}

	withdrawAddr, err := sdk.AccAddressFromBech32(withdrawAddress)
	if err != nil {
		return nil, sdk.Wrapf("%s invalid address", withdrawAddress)
	}
	msg := MsgSetWithdrawAddress{
		Provider:        provider,
		WithdrawAddress: withdrawAddr,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// RefundServiceDeposit refunds the deposit from the specified service binding
func (s serviceClient) RefundServiceDeposit(serviceName string, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}
	msg := MsgRefundServiceDeposit{
		ServiceName: serviceName,
		Provider:    provider,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// StartRequestContext starts the specified request context
func (s serviceClient) StartRequestContext(requestContextID string, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	consumer, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}
	msg := MsgStartRequestContext{
		RequestContextID: rpc.RequestContextIDToByte(requestContextID),
		Consumer:         consumer,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// PauseRequestContext suspends the specified request context
func (s serviceClient) PauseRequestContext(requestContextID string, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	consumer, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}
	msg := MsgPauseRequestContext{
		RequestContextID: rpc.RequestContextIDToByte(requestContextID),
		Consumer:         consumer,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// KillRequestContext terminates the specified request context
func (s serviceClient) KillRequestContext(requestContextID string, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	consumer, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}
	msg := MsgKillRequestContext{
		RequestContextID: rpc.RequestContextIDToByte(requestContextID),
		Consumer:         consumer,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// UpdateRequestContext updates the specified request context
func (s serviceClient) UpdateRequestContext(request rpc.UpdateContextRequest, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	consumer, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}

	var providers []sdk.AccAddress
	for _, provider := range request.Providers {
		p, err := sdk.AccAddressFromBech32(provider)
		if err != nil {
			return nil, sdk.Wrap(err)
		}
		providers = append(providers, p)
	}

	msg := MsgUpdateRequestContext{
		RequestContextID:  rpc.RequestContextIDToByte(request.RequestContextID),
		Providers:         providers,
		ServiceFeeCap:     request.ServiceFeeCap,
		Timeout:           request.Timeout,
		RepeatedFrequency: request.RepeatedFrequency,
		RepeatedTotal:     request.RepeatedTotal,
		Consumer:          consumer,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// WithdrawEarnedFees withdraws the earned fees to the specified provider
func (s serviceClient) WithdrawEarnedFees(baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}

	msg := MsgWithdrawEarnedFees{
		Provider: provider,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

// WithdrawTax withdraws the service tax to the speicified destination address by the trustee
func (s serviceClient) WithdrawTax(destAddress string, amount sdk.Coins, baseTx sdk.BaseTx) (sdk.Result, sdk.Error) {
	trustee, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return nil, sdk.Wrap(err)
	}

	receipt, err := sdk.AccAddressFromBech32(destAddress)
	if err != nil {
		return nil, sdk.Wrapf("%s invalid address", destAddress)
	}
	msg := MsgWithdrawTax{
		Trustee:     trustee,
		DestAddress: receipt,
		Amount:      amount,
	}
	return s.BuildAndSend([]sdk.Msg{msg}, baseTx)
}

//RegisterServiceListener is responsible for registering a group of service handler
func (s serviceClient) RegisterServiceListener(serviceRouter rpc.ServiceRouter, baseTx sdk.BaseTx) sdk.Error {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return sdk.Wrap(err)
	}

	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	builder := sdk.NewEventQueryBuilder().
		AddCondition(sdk.EventKey(tagProvider), sdk.EventValue(provider.String()))
	_, err = s.SubscribeNewBlockWithQuery(builder, func(block sdk.EventDataNewBlock) {
		s.Debug().Int64("height", block.Block.Height).Msg("received block")
		reqIDs := block.ResultEndBlock.Tags.GetValues(tagRequestID)
		for _, reqID := range reqIDs {
			request, err := s.QueryRequest(reqID)
			if !err.IsNil() {
				s.Err(err).Str("requestID", reqID).Msg("service request don't exist")
				continue
			}
			if handler, ok := serviceRouter[request.ServiceName]; ok && provider.Equals(request.Provider) {
				output, errMsg := handler(request.Input)
				msg := MsgRespondService{
					RequestID: reqID,
					Provider:  provider,
					Output:    output,
					Error:     errMsg,
				}
				go func() {
					if _, err = s.BuildAndSend([]sdk.Msg{msg}, baseTx); !err.IsNil() {
						s.Err(err).Str("requestID", reqID).Msg("provider respond failed")
					}
				}()
			}
		}
	})
	return sdk.Wrap(err)
}

//RegisterSingleServiceListener is responsible for registering a single service handler
func (s serviceClient) RegisterSingleServiceListener(serviceName string,
	respondHandler rpc.ServiceRespondHandler,
	baseTx sdk.BaseTx) sdk.Error {
	provider, err := s.QueryAddress(baseTx.From, baseTx.Password)
	if err != nil {
		return sdk.Wrap(err)
	}

	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	// TODO user will don't received any event from tendermint when block_result has the same key in tag
	//builder := sdk.NewEventQueryBuilder().
	//	AddCondition(sdk.EventKey(tagProvider), sdk.EventValue(provider.String())).
	//	AddCondition(sdk.EventKey(tagServiceName), sdk.EventValue(serviceName))
	_, err = s.SubscribeNewBlockWithQuery(nil, func(block sdk.EventDataNewBlock) {
		reqIDs := block.ResultEndBlock.Tags.GetValues(tagRequestID)
		s.Debug().Int64("height", block.Block.Height).Msg("received block")
		for _, reqID := range reqIDs {
			request, err := s.QueryRequest(reqID)
			if !err.IsNil() {
				s.Err(err).Str("requestID", reqID).Msg("service request don't exist")
				continue
			}
			if provider.Equals(request.Provider) && request.ServiceName == serviceName {
				output, errMsg := respondHandler(request.Input)
				msg := MsgRespondService{
					RequestID: reqID,
					Provider:  provider,
					Output:    output,
					Error:     errMsg,
				}
				go func() {
					if _, err = s.BuildAndSend([]sdk.Msg{msg}, baseTx); !err.IsNil() {
						s.Err(err).Str("requestID", reqID).Msg("provider respond failed")
					}
				}()
			}
		}
	})
	return sdk.Wrap(err)
}

// QueryDefinition return a service definition of the specified name
func (s serviceClient) QueryDefinition(serviceName string) (rpc.ServiceDefinition, sdk.Error) {
	param := struct {
		ServiceName string
	}{
		ServiceName: serviceName,
	}

	var definition serviceDefinition
	if err := s.QueryWithResponse("custom/service/definition", param, &definition); err != nil {
		return rpc.ServiceDefinition{}, sdk.Wrap(err)
	}
	return definition.Convert().(rpc.ServiceDefinition), sdk.Nil
}

// QueryBinding return the specified service binding
func (s serviceClient) QueryBinding(serviceName string, provider sdk.AccAddress) (rpc.ServiceBinding, sdk.Error) {
	param := struct {
		ServiceName string
		Provider    sdk.AccAddress
	}{
		ServiceName: serviceName,
		Provider:    provider,
	}

	var binding serviceBinding
	if err := s.QueryWithResponse("custom/service/binding", param, &binding); err != nil {
		return rpc.ServiceBinding{}, sdk.Wrap(err)
	}
	return binding.Convert().(rpc.ServiceBinding), sdk.Nil
}

// QueryBindings returns all bindings of the specified service
func (s serviceClient) QueryBindings(serviceName string) ([]rpc.ServiceBinding, sdk.Error) {
	param := struct {
		ServiceName string
	}{
		ServiceName: serviceName,
	}

	var bindings serviceBindings
	if err := s.QueryWithResponse("custom/service/bindings", param, &bindings); err != nil {
		return nil, sdk.Wrap(err)
	}
	return bindings.Convert().([]rpc.ServiceBinding), sdk.Nil
}

// QueryRequest returns  the active request of the specified requestID
func (s serviceClient) QueryRequest(requestID string) (rpc.RequestService, sdk.Error) {
	param := struct {
		RequestID string
	}{
		RequestID: requestID,
	}

	var request request
	if err := s.QueryWithResponse("custom/service/request", param, &request); err != nil {
		return rpc.RequestService{}, sdk.Wrap(err)
	}
	return request.Convert().(rpc.RequestService), sdk.Nil
}

// QueryRequest returns all the active requests of the specified service binding
func (s serviceClient) QueryRequests(serviceName string, provider sdk.AccAddress) ([]rpc.RequestService, sdk.Error) {
	param := struct {
		ServiceName string
		Provider    sdk.AccAddress
	}{
		ServiceName: serviceName,
		Provider:    provider,
	}

	var rs requests
	if err := s.QueryWithResponse("custom/service/requests", param, &rs); err != nil {
		return nil, sdk.Wrap(err)
	}
	return rs.Convert().([]rpc.RequestService), sdk.Nil
}

// QueryRequestsByReqCtx returns all requests of the specified request context ID and batch counter
func (s serviceClient) QueryRequestsByReqCtx(requestContextID string, batchCounter uint64) ([]rpc.RequestService, sdk.Error) {
	param := struct {
		RequestContextID []byte
		BatchCounter     uint64
	}{
		RequestContextID: rpc.RequestContextIDToByte(requestContextID),
		BatchCounter:     batchCounter,
	}

	var rs requests
	if err := s.QueryWithResponse("custom/service/requests_by_ctx", param, &rs); err != nil {
		return nil, sdk.Wrap(err)
	}
	return rs.Convert().([]rpc.RequestService), sdk.Nil
}

// QueryResponse returns a response with the speicified request ID
func (s serviceClient) QueryResponse(requestID string) (rpc.ServiceResponse, sdk.Error) {
	param := struct {
		RequestID string
	}{
		RequestID: requestID,
	}

	var response response
	if err := s.QueryWithResponse("custom/service/response", param, &response); err != nil {
		return rpc.ServiceResponse{}, sdk.Wrap(nil)
	}
	return response.Convert().(rpc.ServiceResponse), sdk.Nil
}

// QueryResponses returns all responses of the specified request context and batch counter
func (s serviceClient) QueryResponses(requestContextID string, batchCounter uint64) ([]rpc.ServiceResponse, sdk.Error) {
	param := struct {
		RequestContextID []byte
		BatchCounter     uint64
	}{
		RequestContextID: rpc.RequestContextIDToByte(requestContextID),
		BatchCounter:     batchCounter,
	}
	var rs responses
	if err := s.QueryWithResponse("custom/service/responses", param, &rs); err != nil {
		return nil, sdk.Wrap(err)
	}
	return rs.Convert().([]rpc.ServiceResponse), sdk.Nil
}

// QueryRequestContext return the specified request context
func (s serviceClient) QueryRequestContext(requestContextID string) (rpc.RequestContext, sdk.Error) {
	param := struct {
		RequestContextID []byte
	}{
		RequestContextID: rpc.RequestContextIDToByte(requestContextID),
	}

	var reqCtx requestContext
	if err := s.QueryWithResponse("custom/service/context", param, &reqCtx); err != nil {
		return rpc.RequestContext{}, sdk.Wrap(err)
	}
	return reqCtx.Convert().(rpc.RequestContext), sdk.Nil
}

//QueryFees return the earned fees for a provider
func (s serviceClient) QueryFees(provider string) (rpc.EarnedFees, sdk.Error) {
	address, err := sdk.AccAddressFromBech32(provider)
	if err != nil {
		return rpc.EarnedFees{}, sdk.Wrap(err)
	}

	param := struct {
		Address sdk.AccAddress
	}{
		Address: address,
	}

	var fee earnedFees

	if err := s.QueryWithResponse("custom/service/fees", param, &fee); err != nil {
		return rpc.EarnedFees{}, sdk.Wrap(err)
	}
	return fee.Convert().(rpc.EarnedFees), sdk.Nil
}
