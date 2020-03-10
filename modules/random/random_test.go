package random_test

import (
	"testing"

	"github.com/irisnet/irishub-sdk-go/rpc"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/irisnet/irishub-sdk-go/sim"
	sdk "github.com/irisnet/irishub-sdk-go/types"
)

type RandomTestSuite struct {
	suite.Suite
	sim.TestClient
}

func TestSlashingTestSuite(t *testing.T) {
	suite.Run(t, new(RandomTestSuite))
}

func (rts *RandomTestSuite) SetupTest() {
	tc := sim.NewClient()
	rts.TestClient = tc
}

func (rts *RandomTestSuite) TestGenerate() {
	baseTx := sdk.BaseTx{
		From: "test1",
		Gas:  20000,
		Memo: "test",
		Mode: sdk.Commit,
	}

	var memory = make(map[string]string, 1)
	var signal = make(chan int, 0)
	request := rpc.RandomRequest{
		BaseTx:        baseTx,
		BlockInterval: 2,
		Callback: func(reqID, randomNum string, err sdk.Error) {
			require.True(rts.T(), err.IsNil())
			memory[reqID] = randomNum
			signal <- 1
		},
	}
	reqID, err := rts.Random().Generate(request)
	require.True(rts.T(), err.IsNil())
	memory[reqID] = ""
	<-signal
	require.NotEmpty(rts.T(), memory[reqID])
}
