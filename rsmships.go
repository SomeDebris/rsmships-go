package rsmships

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

// command structure. Only stores command flags (e.g. ALWAYS_MANEUVER, AI_BINDING,
// etc.) and faction number.
type CommandData struct {
	// Command flags change the behavior of a ship's AI. Tournament mode
	// overwrites all command flags except for ALWAYS_MANEUVER, ALWAYS_KITE,
	// ALWAYS_RUSH, and AI_BINDING.
	//
	// This is stored as a json.RawMessage, as Reassembly will serialize this
	// feld as a single string instead of a string array when only one flag is
	// specified.
	Flags   json.RawMessage `json:"flags,omitempty"`
	// The faction the command belongs to. Tournament mode overwrites this
	// value to 100, 101, 102, etc.
	Faction int             `json:"faction,omitempty"`
}

// Block structure. Stores basic information needed for Tournaments. Because not
// all common block fields are present, unmarshalling a block into this struct
// will purge all fields that may change the properties of a block. This is
// intentional, as modifying blocks is not illegal in most Reassembly
// Tournaments.
type Block struct {
	// Unqique block ID. Used to identify the block used by the ship.
	//
	// Block IDs are stored as json.RawMessage, as Reassembly will sometimes
	// serialize integers as hexadecimal values. This does not conform to the JSON
	// standard, and, as such, cannot be unmarshalled by encoding/json.
	Id        json.RawMessage `json:"ident"`
	// distance in X and Y between the ship's origin (usually set to its center of mass on
	// export) and the centroid of the block.
	Offset    [2]float64      `json:"offset"`
	Angle     float64         `json:"angle"`
	Command   *CommandData    `json:"command,omitempty"`
	BindingId int             `json:"bindingId,omitempty"`
}

// Defines the data field of a ship blueprint. This contains the ship's name,
// author, colors, and wgroup setting.
type ShipData struct {
	// The name of the ship
	Name   string          `json:"name"`
	// The name of the ship's creator
	Author string          `json:"author"`
	// Ship primary color
	Color0 json.RawMessage `json:"color0,omitempty"`
	// Ship secondary color
	Color1 json.RawMessage `json:"color1,omitempty"`
	// Ship tertiary color
	Color2 json.RawMessage `json:"color2,omitempty"`
	// Weapon binding group setting.
	// Each index specifies whether the weapon group is set to "Fire All" or
	// "Ripple Fire".
	// - [0]: Primary
	// - [1]: Secondary
	// - [2]: Tertiary
	// - [3]: Autofire
	// The values mean this:
	// - if value is 0: set associated binding group to a default value (Fire All)
	// - if value is 1: set associated binding group to Fire All (Fire all
	// weapons at their fire rate)
	// - if value is 2: set associated binding group to Ripple Fire (Fire
	// weapons sequentially with the goal of achieving the maximum possible fire
	// rate. Usually reduces fire rate significantly)
	Wgroup [4]int          `json:"wgroup,omitempty"`
}

// Defines a ship. Marhsal/unmarshal ship files with this datatype.
type Ship struct {
	Angle    float64    `json:"angle,omitempty"`
	Position [2]float64 `json:"position,omitempty"`
	Data     ShipData   `json:"data"`
	Blocks   []Block    `json:"blocks"`
}

// Defines a fleet of multiple ships. This datatype is designed to be used in
// Tournaments. Not intended to store fleets exported from campaign mode.
type Fleet struct {
	// List of ships that the fleet comprises, stored as Ship structures.
	Blueprints []Ship `json:"blueprints"`
	// Fleet primary color
	Color0     any    `json:"color0,omitempty"`
	// Fleet secondary color
	Color1     any    `json:"color1,omitempty"`
	// Fleet tertiary color
	Color2     any    `json:"color2,omitempty"`
	// Faction number of fleet. All commands will be assigned to this faction on
	// import into sandbox. Overwritten in Tournament mode.
	Faction    int    `json:"faction"`
	// The name of the fleet
	Name       string `json:"name"`
}

// A dummy datatype to unmarshal data to to determine whether the data is a Ship
// file or a Fleet file
type UnprocessedShip struct {
	// The name of the fleet. If this field exists, the data defines a Fleet
	// file. If this field does not exist, the data defines a different type of
	// file.
	Name json.RawMessage `json:"name"`
}

func IsReassemblyJSONFileFleet(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	var idk UnprocessedShip

	if err := json.Unmarshal([]byte(content), &idk); err != nil {
		return false, err
	}

	if idk.Name == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func UnmarshalShipFromFile(path string) (Ship, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Ship{}, err
	}

	var ship Ship

	if err := json.Unmarshal([]byte(content), &ship); err != nil {
		return Ship{}, err
	}

	return RemoveNilIds(ship), nil
}

func RemoveNilIds(ship Ship) Ship {
	// new_blocks := make([]blocks
	blocks := 0

	for _, block := range ship.Blocks {
		if block.Id != nil {
			blocks++
		}
	}

	new_blocks := make([]Block, blocks)
	block_idx := 0
	for _, block := range ship.Blocks {
		if block.Id != nil {
			new_blocks[block_idx] = block
			block_idx++
		}
	}

	ship.Blocks = new_blocks

	return ship
}

func MarshalShipToFile(path string, ship Ship) error {
	b, err := json.Marshal(ship)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, b, 0666); err != nil {
		return err
	}

	return nil
}

func UnmarshalFleetFromFile(path string) (Fleet, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Fleet{}, err
	}

	var fleet Fleet

	if err := json.Unmarshal([]byte(content), &fleet); err != nil {
		return Fleet{}, err
	}

	return fleet, nil
}

func MarshalFleetToFile(path string, fleet Fleet) error {
	b, err := json.Marshal(fleet)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, _ := gzip.NewWriterLevel(file, gzip.BestCompression)
	gz.Write(b)
	gz.Close()

	return nil
}

func FleetFromShips(template Fleet, ships ...Ship) Fleet {
	template.Blueprints = ships

	return template
}

func AssembleAlliance(template Fleet, ships []Ship) Fleet {
	template.Blueprints = ships

	return template
}
