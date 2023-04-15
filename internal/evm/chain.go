package evm

import (
	"embed"
	"encoding/json"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/utils"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"
)

type InitState int

var currentState InitState
var initStateMutex sync.RWMutex

const (
	InitStateIdle = InitState(iota)
	InitStateStarted
	InitStateFinished
)

var ()

//go:embed chains
var chainsDir embed.FS
var chainsDirName = "chains"

//go:embed iconsDownload
var iconsDownloadDir embed.FS
var iconsDownloadDirName = "iconsDownload"

//go:embed icons
var iconsDir embed.FS
var iconsDirName = "icons"

var chainsSlice []Chain
var chainsIDMap = utils.NewMap[string, Chain]()

type Chain struct {
	Name           string         `json:"name"`
	Chain          string         `json:"chain"`
	Icon           string         `json:"icon,omitempty"`
	IconData       []IconData     `json:",omitempty"`
	IconImage      []byte         `json:",omitempty"`
	RPC            []RPC          `json:"rpc"`
	Features       []Features     `json:"features,omitempty"`
	Faucets        []string       `json:"faucets"`
	NativeCurrency NativeCurrency `json:"nativeCurrency"`
	InfoURL        string         `json:"infoURL"`
	ShortName      string         `json:"shortName"`
	ChainID        big.Int        `json:"chainId"`
	NetworkID      big.Int        `json:"networkId"`
	Slip44         big.Int        `json:"slip44,omitempty"`
	Ens            Ens            `json:"ens,omitempty"`
	Explorers      []Explorer     `json:"explorers,omitempty"`
	Title          string         `json:"title,omitempty"`
	RedFlags       []string       `json:"redFlags,omitempty"`
	Parent         Parent         `json:"parent,omitempty"`
	Status         string         `json:"status,omitempty"`
}

type Features struct {
	Name string `json:"name"`
}
type NativeCurrency struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

type Ens struct {
	Registry string `json:"registry"`
}
type Explorer struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Standard string `json:"standard"`
}
type Parent struct {
	Type    string   `json:"type"`
	Chain   string   `json:"chain"`
	Bridges []Bridge `json:"bridges"`
}
type Bridge struct {
	URL string `json:"url"`
}
type IconMetadata struct {
	URL    string     `json:"url"`
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Format IconFormat `json:"format"`
}

type IconData struct {
	IconMetadata `json:",omitempty"`
	Data         []byte `json:",omitempty"`
}
type IconImage []byte

type IconFormat string

const (
	IconFormatPNG = IconFormat("png")
	IconFormatJPG = IconFormat("jpg")
	IconFormatSVG = IconFormat("svg")
)

// iconDataMap is a map[Chain.Icon][]IconData
// Chain.Icon maps to filenames inside icons-metadata(excluding .json extension)
// []IconData is derived from the contents of the corresponding(matched filename) file
var iconDataMap = utils.NewMap[string, []IconData]()

// iconNameIconData map
// string is hashed image name and []byte is encoded image data from icons folder
var iconNameImageMap = utils.NewMap[string, []byte]()

func init() {
	setInitState(InitStateStarted)
	go func() {
		defer func() {
			setInitState(InitStateFinished)
		}()
		files, err := chainsDir.ReadDir(chainsDirName)
		if err != nil {
			alog.Logger().Errorln(err)
		}
		for _, file := range files {
			var val []byte
			val, err = chainsDir.ReadFile(chainsDirName + "/" + file.Name())
			if err != nil {
				continue
			}
			var chain Chain
			err = json.Unmarshal(val, &chain)
			if err != nil {
				continue
			}
			chainsSlice = append(chainsSlice, chain)
		}
		files, err = iconsDownloadDir.ReadDir(iconsDownloadDirName)
		if err != nil {
			alog.Logger().Errorln(err)
		}
		for _, file := range files {
			var val []byte
			val, err = iconsDownloadDir.ReadFile(iconsDownloadDirName + "/" + file.Name())
			if err != nil {
				continue
			}
			iconNameSlice := strings.Split(file.Name(), ".json")
			iconName := strings.Join(iconNameSlice, "")
			var iconDatas []IconData
			err = json.Unmarshal(val, &iconDatas)
			if err != nil {
				continue
			}
			for _, iconData := range iconDatas {
				if iconData.URL != "" {
					imageNameSlice := strings.Split(iconData.URL, "/")
					imageName := strings.Join(imageNameSlice[len(imageNameSlice)-1:], "")
					iconNameImageMap.Set(imageName, []byte{})
				}
			}
			iconDataMap.Set(iconName, iconDatas)
		}
		files, err = iconsDir.ReadDir(iconsDirName)
		if err != nil {
			alog.Logger().Errorln(err)
		}
		for _, file := range files {
			var val []byte
			val, err = iconsDir.ReadFile(iconsDirName + "/" + file.Name())
			if err != nil {
				continue
			}
			if _, ok := iconNameImageMap.Get(file.Name()); ok {
				iconNameImageMap.Set(file.Name(), val)
			}
		}
		for i, chain := range chainsSlice {
			iconDatas, ok := iconDataMap.Get(chainsSlice[i].Icon)
			if ok {
				for j, iconData := range iconDatas {
					if iconData.URL != "" {
						imageNameSlice := strings.Split(iconData.URL, "/")
						imageName := strings.Join(imageNameSlice[len(imageNameSlice)-1:], "")
						if val, ok := iconNameImageMap.Get(imageName); ok {
							iconDatas[j].Data = val
						}
					}
				}
				chainsSlice[i].IconData = iconDatas
			}
			chainsIDMap.Set(chainsSlice[i].ChainID.String(), chain)
		}
		sort.Slice(chainsSlice, func(i, j int) bool {
			return strings.ToLower(chainsSlice[i].Name) < strings.ToLower(chainsSlice[j].Name)
		})
	}()
}

func getCurrentState() InitState {
	initStateMutex.RLock()
	defer initStateMutex.RUnlock()
	return currentState
}
func ChainsSlice() []Chain {
	for getCurrentState() < InitStateFinished {
		time.Sleep(0)
	}
	return chainsSlice
}

func setInitState(state InitState) {
	initStateMutex.Lock()
	currentState = state
	initStateMutex.Unlock()
}
