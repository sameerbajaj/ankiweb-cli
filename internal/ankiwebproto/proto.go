// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.

package ankiwebproto

import (
	"context"
	"fmt"
	"strings"

	"github.com/sameerbajaj/ankiweb-cli/internal/client"
)

type PBField struct {
	Num      int
	WireType int
	Varint   uint64
	Bytes    []byte
}

func EncodeVarint(val uint64) []byte {
	var b []byte
	for {
		c := byte(val & 0x7f)
		val >>= 7
		if val != 0 {
			b = append(b, c|0x80)
		} else {
			b = append(b, c)
			break
		}
	}
	return b
}

func EncodeString(fieldNum int, s string) []byte {
	data := []byte(s)
	tag := EncodeVarint(uint64((fieldNum << 3) | 2))
	length := EncodeVarint(uint64(len(data)))
	res := append(tag, length...)
	return append(res, data...)
}

func EncodeMessage(fieldNum int, msg []byte) []byte {
	tag := EncodeVarint(uint64((fieldNum << 3) | 2))
	length := EncodeVarint(uint64(len(msg)))
	res := append(tag, length...)
	return append(res, msg...)
}

func ParsePB(data []byte) []PBField {
	var fields []PBField
	i := 0
	for i < len(data) {
		tag := data[i]
		fieldNum := int(tag >> 3)
		wireType := int(tag & 7)
		i++
		if wireType == 0 {
			var val uint64
			var shift uint
			for i < len(data) {
				b := data[i]
				i++
				val |= uint64(b&0x7f) << shift
				shift += 7
				if (b & 0x80) == 0 {
					break
				}
			}
			fields = append(fields, PBField{Num: fieldNum, WireType: wireType, Varint: val})
		} else if wireType == 2 {
			var length uint64
			var shift uint
			for i < len(data) {
				b := data[i]
				i++
				length |= uint64(b&0x7f) << shift
				shift += 7
				if (b & 0x80) == 0 {
					break
				}
			}
			end := i + int(length)
			if end > len(data) {
				end = len(data)
			}
			val := data[i:end]
			i = end
			fields = append(fields, PBField{Num: fieldNum, WireType: wireType, Bytes: val})
		} else {
			// Skip unrecognized wire types safely
			break
		}
	}
	return fields
}

type NotetypeInfo struct {
	ID     uint64   `json:"id"`
	Name   string   `json:"name"`
	Fields []string `json:"fields,omitempty"`
}

type DeckInfo struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type AddInfo struct {
	Notetypes         []NotetypeInfo `json:"notetypes"`
	Decks             []DeckInfo     `json:"decks"`
	DefaultDeckID     uint64         `json:"default_deck_id"`
	DefaultNotetypeID uint64         `json:"default_notetype_id"`
	DefaultFields     []string       `json:"default_fields"`
}

func FetchAddInfo(ctx context.Context, c *client.Client) (*AddInfo, error) {
	data, statusCode, err := c.PostProtobufQuery(ctx, "/svc/editor/get-info-for-adding", []byte{})
	if err != nil {
		return nil, fmt.Errorf("fetch get-info-for-adding: %w", err)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("fetch get-info-for-adding returned status %d", statusCode)
	}

	info := &AddInfo{}

	top := ParsePB(data)
	for _, f := range top {
		switch f.Num {
		case 1: // notetypes
			sub := ParsePB(f.Bytes)
			var nt NotetypeInfo
			for _, sf := range sub {
				if sf.Num == 1 && sf.WireType == 0 {
					nt.ID = sf.Varint
				} else if sf.Num == 2 && sf.WireType == 2 {
					nt.Name = string(sf.Bytes)
				}
			}
			if nt.ID != 0 {
				info.Notetypes = append(info.Notetypes, nt)
			}
		case 2: // decks
			sub := ParsePB(f.Bytes)
			var d DeckInfo
			for _, sf := range sub {
				if sf.Num == 1 && sf.WireType == 0 {
					d.ID = sf.Varint
				} else if sf.Num == 2 && sf.WireType == 2 {
					d.Name = string(sf.Bytes)
				}
			}
			if d.ID != 0 {
				info.Decks = append(info.Decks, d)
			}
		case 3: // default deck id
			if f.WireType == 0 && f.Varint != 0 {
				info.DefaultDeckID = f.Varint
			}
		case 4: // default notetype id
			if f.WireType == 0 && f.Varint != 0 {
				info.DefaultNotetypeID = f.Varint
			}
		case 5: // default notetype fields
			sub := ParsePB(f.Bytes)
			for _, sf := range sub {
				if sf.Num == 2 && sf.WireType == 2 {
					name := string(sf.Bytes)
					if name != "" {
						info.DefaultFields = append(info.DefaultFields, name)
					}
				}
			}
		}
	}

	if info.DefaultDeckID == 0 {
		info.DefaultDeckID = 1
	}
	if info.DefaultNotetypeID == 0 {
		info.DefaultNotetypeID = 1731906299829
	}
	if len(info.DefaultFields) == 0 {
		info.DefaultFields = []string{"Front", "Back", "Additional Info", "Source"}
	}

	return info, nil
}

func FetchNotetypeFields(ctx context.Context, c *client.Client, notetypeID uint64) ([]string, error) {
	reqBody := append(EncodeVarint(uint64(1<<3)), EncodeVarint(notetypeID)...)
	data, statusCode, err := c.PostProtobufQuery(ctx, "/svc/editor/get-notetype-fields", reqBody)
	if err != nil {
		return nil, fmt.Errorf("fetch get-notetype-fields: %w", err)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("fetch get-notetype-fields returned status %d", statusCode)
	}

	var fields []string
	top := ParsePB(data)
	for _, f := range top {
		if f.Num == 1 && f.WireType == 2 {
			sub := ParsePB(f.Bytes)
			for _, sf := range sub {
				if sf.Num == 2 && sf.WireType == 2 {
					name := string(sf.Bytes)
					if name != "" {
						fields = append(fields, name)
					}
				}
			}
		}
	}
	return fields, nil
}

type AddCardRequest struct {
	DeckName     string
	NotetypeName string
	Front        string
	Back         string
	Fields       map[string]string
	Tags         []string
}

type AddCardResponse struct {
	Success    bool              `json:"success"`
	DeckID     uint64            `json:"deck_id"`
	DeckName   string            `json:"deck"`
	NotetypeID uint64            `json:"notetype_id"`
	Notetype   string            `json:"notetype"`
	Fields     map[string]string `json:"fields"`
	Tags       []string          `json:"tags"`
}

func AddCard(ctx context.Context, c *client.Client, req AddCardRequest) (*AddCardResponse, error) {
	info, err := FetchAddInfo(ctx, c)
	if err != nil {
		// Use standard defaults if offline/dry-run/mock
		info = &AddInfo{
			DefaultDeckID:     1,
			DefaultNotetypeID: 1731906299829,
			DefaultFields:     []string{"Front", "Back", "Additional Info", "Source"},
		}
	}

	deckID := info.DefaultDeckID
	targetDeck := req.DeckName
	if targetDeck == "" {
		targetDeck = "Default"
	}
	for _, d := range info.Decks {
		if strings.EqualFold(d.Name, targetDeck) {
			deckID = d.ID
			targetDeck = d.Name
			break
		}
	}

	notetypeID := info.DefaultNotetypeID
	targetNotetype := req.NotetypeName
	if targetNotetype == "" {
		targetNotetype = "Basic"
	}
	for _, nt := range info.Notetypes {
		if strings.EqualFold(nt.Name, targetNotetype) {
			notetypeID = nt.ID
			targetNotetype = nt.Name
			break
		}
	}

	// Resolve field definitions for target notetype
	var fieldDefs []string
	if notetypeID == info.DefaultNotetypeID && len(info.DefaultFields) > 0 {
		fieldDefs = info.DefaultFields
	} else {
		fields, fErr := FetchNotetypeFields(ctx, c, notetypeID)
		if fErr == nil && len(fields) > 0 {
			fieldDefs = fields
		} else {
			fieldDefs = []string{"Front", "Back", "Additional Info", "Source"}
		}
	}

	// Build field values matching field definitions in order
	assignedFields := make(map[string]string)
	var orderedValues []string

	for i, fName := range fieldDefs {
		val := ""
		lower := strings.ToLower(strings.TrimSpace(fName))

		// Check explicit custom fields first
		for k, v := range req.Fields {
			if strings.EqualFold(k, fName) {
				val = v
				break
			}
		}

		// Check primary positional/flag mappings if not matched
		if val == "" {
			if lower == "front" || lower == "text" || i == 0 {
				val = req.Front
			} else if lower == "back" || lower == "back extra" || i == 1 {
				val = req.Back
			} else if lower == "additional info" || lower == "extra" {
				if v, ok := req.Fields["additional_info"]; ok {
					val = v
				} else if v, ok := req.Fields["extra"]; ok {
					val = v
				}
			} else if lower == "source" {
				if v, ok := req.Fields["source"]; ok {
					val = v
				}
			}
		}

		assignedFields[fName] = val
		orderedValues = append(orderedValues, val)
	}

	// Construct Protobuf binary payload
	addMsg := append(EncodeTag(1, 0), EncodeVarint(notetypeID)...)
	addMsg = append(addMsg, append(EncodeTag(2, 0), EncodeVarint(deckID)...)...)

	var reqBytes []byte
	// Field 1: repeated field values in order
	for _, val := range orderedValues {
		reqBytes = append(reqBytes, EncodeString(1, val)...)
	}
	// Field 2: tags string
	tagsStr := strings.Join(req.Tags, " ")
	reqBytes = append(reqBytes, EncodeString(2, tagsStr)...)
	// Field 3: add mode submessage
	reqBytes = append(reqBytes, EncodeMessage(3, addMsg)...)

	_, statusCode, postErr := c.PostProtobuf(ctx, "/svc/editor/add-or-update", reqBytes)
	if postErr != nil {
		return nil, fmt.Errorf("add card request failed: %w", postErr)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("add card returned status %d", statusCode)
	}

	return &AddCardResponse{
		Success:    true,
		DeckID:     deckID,
		DeckName:   targetDeck,
		NotetypeID: notetypeID,
		Notetype:   targetNotetype,
		Fields:     assignedFields,
		Tags:       req.Tags,
	}, nil
}

func EncodeTag(fieldNum int, wireType int) []byte {
	return EncodeVarint(uint64((fieldNum << 3) | wireType))
}
