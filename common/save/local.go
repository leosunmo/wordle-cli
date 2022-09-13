package save

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type LocalStorage struct{}

func (ls LocalStorage) Load(id string, user uint64) (*SaveFile, error) {
	savepath, err := ls.getSaveLocation(id)
	if err != nil {
		return nil, err
	}

	return ls.loadSave(savepath)
}

func (ls LocalStorage) Save(save *SaveFile, id string, user uint64) error {
	savepath, err := ls.getSaveLocation(id)
	if err != nil {
		return err
	}

	data, err := json.Marshal(save)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(savepath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (ls LocalStorage) loadSave(savepath string) (*SaveFile, error) {
	data, err := ioutil.ReadFile(savepath)
	if err != nil {
		return nil, err
	}

	save := NewSave()
	err = json.Unmarshal(data, save)

	if err != nil {
		return nil, err
	}

	return save, nil
}

func (ls LocalStorage) getSaveLocation(id string) (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/.wordlecli_%s.save.json", dir, id), nil
}
