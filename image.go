package bus_tracker

import (
	"bytes"
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/google/uuid"
	storage_go "github.com/supabase-community/storage-go"
)

var StorageClient *storage_go.Client

type Image struct {
	Url         string
	Name        string
	ContentType string
	Body        []byte
}

func NewImageInstance(image *Image) *lox.LoxInstance {
	instance := lox.NewLoxInstance(
		lox.NewLoxClass("Image", nil, map[string]lox.Callable{
			"save": newImageFunction("save", 0, save),
		}))

	_ = instance.Set(lox.Token{Lexeme: "_image"}, lox.NewLiteralExpr(image))

	return instance
}

type imageFunctionCall func(image *Image, arguments []any) (v interface{}, err error)

var _ lox.Callable = (*ImageFunction)(nil)

type ImageFunction struct {
	instance *lox.LoxInstance
	arity    int
	call     imageFunctionCall
	name     string
	imageKey string
}

func newImageFunction(name string, arity int, call imageFunctionCall) *ImageFunction {
	return &ImageFunction{
		arity:    arity,
		call:     call,
		name:     name,
		imageKey: "_image",
	}
}

func (f ImageFunction) Call(i *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	image, err := f.instance.Get(lox.Token{Lexeme: f.imageKey})
	if err != nil {
		return
	}

	_, isLocator := image.(*Image)
	if !isLocator {
		return nil, fmt.Errorf("is not Locator")
	}

	v, err = f.call(image.(*Image), arguments)

	return v, err
}

func (f ImageFunction) Arity() int {
	return f.arity
}

func (f ImageFunction) ToString() string {
	return fmt.Sprintf("<native fn %s>", f.name)
}

func (f ImageFunction) Bind(instance *lox.LoxInstance) lox.Callable {
	f.instance = instance
	return f
}

func save(image *Image, arguments []interface{}) (v interface{}, err error) {
	path := uuid.New().String()
	_, err = StorageClient.UploadFile("images", path, bytes.NewReader(image.Body), storage_go.FileOptions{
		ContentType: &image.ContentType,
	})
	if err != nil {
		return nil, err
	}

	publicUrl := StorageClient.GetPublicUrl("images", path)

	return publicUrl, nil
}
