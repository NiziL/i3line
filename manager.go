package i3line

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type BlockModule interface {
	GetName() string
	GenBlock() Block
	OnClick(Event)
}

type BlockManager struct {
	modules  []BlockModule
	lastSend string
}

func NewBlockManager() *BlockManager {
	manager := new(BlockManager)
	manager.modules = make([]BlockModule, 0)
	manager.lastSend = ""
	return manager
}

func (m *BlockManager) Init() {
	fmt.Println(`{ "version": 1, "click_events": true }`)
	fmt.Println(`[`)
}

func (m *BlockManager) Close() {
	fmt.Println(`]`)
}

func (m *BlockManager) AddBlockModule(module BlockModule) {
	for _, oldModule := range m.modules {
		if oldModule.GetName() == module.GetName() {
			panic("Can't have two modules with the same name")
		}
	}
	m.modules = append(m.modules, module)
}

func (m *BlockManager) Run() {
	go func() {
		//TODO scheduler instead of for loop
		for {
			var blocks []Block
			for _, module := range m.modules {
				blocks = append(blocks, module.GenBlock())
			}
			m.refreshBlocks(blocks)
		}
	}()
	m.listenEvent()
}

func (m *BlockManager) refreshBlocks(blocks []Block) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(blocks)
	if buf.String() != m.lastSend {
		fmt.Println(buf.String() + ",")
		m.lastSend = buf.String()
	}
}

type Event struct {
	Name     string `json: "name"`
	Instance string `json: "instance"`
	Button   int    `json: "button"`
	X        int    `json: "x"`
	Y        int    `json: "y"`
}

func (m *BlockManager) listenEvent() {
	// new json decoder from stdin
	decoder := json.NewDecoder(bufio.NewReader(os.Stdin))

	// read opening bracket
	if _, err := decoder.Token(); err != nil {
		panic(err)
	}

	// read while i3bar sends clickEvent, should be infinite
	var event Event
	for decoder.More() {
		err := decoder.Decode(&event)
		if err != nil {
			panic(err)
		}
		go m.handleEvent(event)
	}

	// read closing bracket
	if _, err := decoder.Token(); err != nil {
		panic(err)
	}
}

func (m *BlockManager) handleEvent(e Event) {
	for _, module := range m.modules {
		if module.GetName() == e.Name {
			module.OnClick(e)
			break
		}
	}
}
