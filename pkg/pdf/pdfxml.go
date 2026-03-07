package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	pdf "github.com/sassoftware/pdf-xtract"
)

type PDF struct {
	FileName string
	Info     map[string]string
}

func NewPDF(path string) (p *PDF, err error) {
	// Защита от паники внутри бинарного парсера
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pdf-xtract panic on file %s: %v", path, r)
			p = nil
		}
	}()

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// Создаем ридер через новый API
	r, err := pdf.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	p = &PDF{
		FileName: filepath.Base(path),
		Info:     make(map[string]string),
	}

	// Извлекаем метаданные из словаря Info
	infoDict := r.Trailer().Key("Info")
	// В pdf-xtract проверка на наличие ключа делается через .IsNull()
	if !infoDict.IsNull() {
		fields := []string{"Title", "Author", "Subject", "CreationDate", "Keywords"}
		for _, field := range fields {
			if val := infoDict.Key(field); !val.IsNull() {
				p.Info[field] = val.String()
			}
		}
	}

	return p, nil
}