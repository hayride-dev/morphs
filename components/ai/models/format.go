package models

import (
	"unsafe"

	"github.com/hayride-dev/morphs/components/ai/models/internal/gen/hayride/ai/model"
	"github.com/hayride-dev/morphs/components/ai/models/internal/gen/hayride/ai/types"
	"go.bytecodealliance.org/cm"
)

type Format interface {
	Decode(data []byte) (types.Message, error)
	Encode(messages ...types.Message) ([]byte, error)
}

type resources struct {
	format map[cm.Rep]Format
}

var resourceTable = &resources{
	format: make(map[cm.Rep]Format),
}

func init() {
	model.Exports.Format.Constructor = constructor
	model.Exports.Format.Decode = decode
	model.Exports.Format.Encode = encode
	model.Exports.Format.Destructor = destructor

}

func constructor() model.Format {
	// This function is called at compile time to create a new format resource.
	// It is used to add in the custom model format and intended to create a new wasm component
	// that can be used to encode and decode messages for a specific model.
	m := comptimeFormat()
	key := cm.Rep(uintptr(*(*unsafe.Pointer)(unsafe.Pointer(&m))))
	v := model.FormatResourceNew(key)
	resourceTable.format[key] = m
	return v
}

func destructor(self cm.Rep) {
	delete(resourceTable.format, self)
}

func decode(self cm.Rep, raw cm.List[uint8]) (result cm.Result[model.MessageShape, model.Message, model.Error]) {
	m, ok := resourceTable.format[self]
	if !ok {
		wasiErr := model.ErrorResourceNew(cm.Rep(model.ErrorCodeContextDecode))
		return cm.Err[cm.Result[model.MessageShape, model.Message, model.Error]](wasiErr)
	}
	msg, err := m.Decode(raw.Slice())
	if err != nil {
		wasiErr := model.ErrorResourceNew(cm.Rep(model.ErrorCodeContextDecode))
		return cm.Err[cm.Result[model.MessageShape, model.Message, model.Error]](wasiErr)
	}

	return cm.OK[cm.Result[model.MessageShape, model.Message, model.Error]](msg)
}

func encode(self cm.Rep, messages cm.List[model.Message]) (result cm.Result[cm.List[uint8], cm.List[uint8], model.Error]) {
	m, ok := resourceTable.format[self]
	if !ok {
		wasiErr := model.ErrorResourceNew(cm.Rep(model.ErrorCodeContextEncode))
		return cm.Err[cm.Result[cm.List[uint8], cm.List[uint8], model.Error]](wasiErr)
	}

	msg, err := m.Encode(messages.Slice()...)
	if err != nil {
		wasiErr := model.ErrorResourceNew(cm.Rep(model.ErrorCodeContextEncode))
		return cm.Err[cm.Result[cm.List[uint8], cm.List[uint8], model.Error]](wasiErr)
	}

	return cm.OK[cm.Result[cm.List[uint8], cm.List[uint8], model.Error]](cm.ToList(msg))
}

func main() {}
