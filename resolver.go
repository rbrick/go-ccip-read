package ccip

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Option func(*CCIPReadResolver)

func Gateways(gateways ...common.Address) Option {
	return func(r *CCIPReadResolver) {
		r.senderValidator = func(sender common.Address) error {
			for _, gw := range gateways {
				if sender == gw {
					return nil
				}
			}
			return nil
		}
	}
}

func GatewayValidator(validator SenderValidator) Option {
	return func(r *CCIPReadResolver) {
		r.senderValidator = validator
	}
}

func OutputEncoding(encoder OutputEncoder) Option {
	return func(r *CCIPReadResolver) {
		r.outputEncoder = encoder
	}
}

type CCIPReadRequest struct {
	Method *abi.Method
	Input  []Variable
}

func (r *CCIPReadRequest) Var(name string) (interface{}, bool) {
	for _, v := range r.Input {
		if v.Name == name {
			return v.Value, true
		}
	}
	return nil, false
}

// CCIPReadHandler defines the function signature for handling CCIP read requests.
type CCIPReadHandler func(request *CCIPReadRequest) ([]interface{}, error)

// SenderValidator defines the function signature for validating the sender address.
type SenderValidator func(sender common.Address) error

// OutputEncoder defines the function signature for encoding output values.
type OutputEncoder func(outputs []interface{}) ([]byte, error)

type registeredHandler struct {
	method  *abi.Method
	handler CCIPReadHandler
}

type CCIPReadResolver struct {
	handlers map[uint32]registeredHandler

	senderValidator SenderValidator

	outputEncoder OutputEncoder
}

// Handle registers a handler for a given function signature.
func (r *CCIPReadResolver) Handle(sig string, handler CCIPReadHandler) error {
	method, err := ParseFunction(sig)

	if err != nil {
		return err
	}

	byte4Sig := binary.BigEndian.Uint32(method.ID)

	r.handlers[byte4Sig] = registeredHandler{
		method:  method,
		handler: handler,
	}

	return nil
}

type HttpCCIPReadRequest struct {
	Data   string `json:"data"`
	Sender string `json:"sender,omitempty"`
}

func (r *CCIPReadResolver) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(rw, "failed to read request body", http.StatusBadRequest)
		return
	}

	var ccipReq HttpCCIPReadRequest

	err = json.Unmarshal(body, &ccipReq)

	if err != nil {
		http.Error(rw, "failed to parse request body", http.StatusBadRequest)
		return
	}

	if r.senderValidator != nil && ccipReq.Sender != "" {
		senderAddr := common.HexToAddress(ccipReq.Sender)

		err := r.senderValidator(senderAddr)

		if err != nil {
			http.Error(rw, "unauthorized sender", http.StatusUnauthorized)
			return
		}
	}

	data, err := hexutil.Decode(ccipReq.Data)

	if err != nil {
		http.Error(rw, "invalid data field", http.StatusBadRequest)
		return
	}
	if len(data) < 4 {
		http.Error(rw, "data field too short", http.StatusBadRequest)
		return
	}

	byte4Sig := binary.BigEndian.Uint32(data[:4])

	registered, ok := r.handlers[byte4Sig]

	if !ok {
		http.Error(rw, "function not found", http.StatusNotFound)
		return
	}

	inputs, err := registered.method.Inputs.UnpackValues(data[4:])

	if err != nil {
		http.Error(rw, "failed to unpack input parameters", http.StatusBadRequest)
		return
	}

	var inputVars []Variable

	for i, input := range inputs {
		inputVars = append(inputVars, Variable{
			Name:  registered.method.Inputs[i].Name,
			Value: input,
		})
	}

	ccipReadReq := &CCIPReadRequest{
		Method: registered.method,
		Input:  inputVars,
	}

	outputs, err := registered.handler(ccipReadReq)

	if err != nil {
		http.Error(rw, "handler error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	outputData, err := registered.method.Outputs.PackValues(outputs)

	if err != nil {
		http.Error(rw, "failed to pack output parameters", http.StatusInternalServerError)
		return
	}

	response := struct {
		Data string `json:"data"`
	}{
		Data: hexutil.Encode(outputData),
	}

	respBytes, err := json.Marshal(response)

	if err != nil {
		http.Error(rw, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(respBytes)
}

func NewCCIPReadResolver(options ...Option) *CCIPReadResolver {
	r := &CCIPReadResolver{
		handlers: make(map[uint32]registeredHandler),
	}

	for _, option := range options {
		option(r)
	}

	return r
}
