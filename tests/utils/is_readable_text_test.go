package utils

import (
	"testing"

	"github.com/mkloubert/go-package-manager/utils"
)

func TestIsReadableTextFunctionWithTextFile(t *testing.T) {
	isText := utils.IsReadableText(test1_txt)
	if !isText {
		t.Error("should be readable text")
	}
}

func TestIsReadableTextFunctionWithImageFile(t *testing.T) {
	isText := utils.IsReadableText(test1_jpg)
	if isText {
		t.Error("should be no readable text")
	}
}
