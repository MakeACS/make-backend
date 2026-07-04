package models

import (
	"encoding/json"
	"fmt"
)

type TrainingBlock interface {
	IsTrainingBlock()
	StripAnswer()
}

// --- BEGIN TRAINING BLOCKS ---
type TextBlock struct {
	BlockId string `json:"block_id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (TextBlock) IsTrainingBlock() {}
func (*TextBlock) StripAnswer()    {}

type ImageBlock struct {
	BlockId string `json:"block_id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (ImageBlock) IsTrainingBlock() {}
func (*ImageBlock) StripAnswer()    {}

type OptionBlockOption struct {
	Text    string `json:"text"`
	Correct *bool  `json:"correct"`
}

type OptionBlock struct {
	BlockId     string              `json:"block_id"`
	Type        string              `json:"type"`
	Content     string              `json:"content"`
	Options     []OptionBlockOption `json:"options"`
	Hint        string              `json:"hint"`
	Affirmation string              `json:"affirmation"`
}

func (OptionBlock) IsTrainingBlock() {}
func (b *OptionBlock) StripAnswer() {
	for i := range b.Options {
		b.Options[i].Correct = nil
	}
}

type ShortAnswerBlock struct {
	BlockId     string  `json:"block_id"`
	Type        string  `json:"type"`
	Content     string  `json:"content"`
	Answer      *string `json:"answer"`
	Hint        string  `json:"hint"`
	Affirmation string  `json:"affirmation"`
}

func (ShortAnswerBlock) IsTrainingBlock() {}
func (b *ShortAnswerBlock) StripAnswer() {
	b.Answer = nil
}

// --- END TRAINING BLOCKS ---
type TrainingBlockList []TrainingBlock

func (list *TrainingBlockList) UnmarshalJSON(data []byte) error {
	var raw_messages []json.RawMessage
	if err := json.Unmarshal(data, &raw_messages); err != nil {
		return err
	}

	*list = make(TrainingBlockList, 0, len(raw_messages))

	for _, raw := range raw_messages {
		var type_checker struct {
			Type string
		}
		if err := json.Unmarshal(raw, &type_checker); err != nil {
			return err
		}

		var block TrainingBlock
		switch type_checker.Type {
		case "TEXT":
			var t TextBlock
			if err := json.Unmarshal(raw, &t); err != nil {
				return err
			}
			block = &t

		case "IMAGE":
			var i ImageBlock
			if err := json.Unmarshal(raw, &i); err != nil {
				return err
			}
			block = &i

		case "MULTI_OPTION", "SINGLE_OPTION":
			var o OptionBlock
			if err := json.Unmarshal(raw, &o); err != nil {
				return err
			}
			block = &o

		case "SHORT_ANSWER":
			var s ShortAnswerBlock
			if err := json.Unmarshal(raw, &s); err != nil {
				return err
			}
			block = &s

		default:
			return fmt.Errorf("Unknown TrainingBlock type: %s", type_checker.Type)
		}

		*list = append(*list, block)
	}

	return nil
}

type Training struct {
	Id           int
	Name         string
	MakerspaceId *int
	Blocks       TrainingBlockList
}

func (b *TrainingBlockList) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected byte for JSONB, got %T", value)
	}

	return json.Unmarshal(bytes, b)
}
