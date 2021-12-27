package random

import (
	"testing"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/gofrs/uuid"
)

func TestSetupMemberRoles(t *testing.T) {
	type args struct {
		g       *model.Game
		members []model.User
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "setup roles",
			args: args{
				g: &model.Game{
					Odais: []*model.Odai{
						{
							Title:      "odai1",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai2",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000002")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai3",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000003")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai4",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000004")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai5",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000005")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai6",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000006")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai7",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000007")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai8",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000008")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai9",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000009")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai10",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000010")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai11",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000011")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai12",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000012")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai13",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000013")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai14",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000014")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						{
							Title:      "odai15",
							SenderId:   model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000015")),
							AnswererId: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq:  []model.Drawer{},
						},
						// {
						// 	Title:     "odai16",
						// 	SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000016")),
						// 	AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
						// 	DrawerSeq: []model.Drawer{},
						// },
						// {
						// 	Title:     "odai17",
						// 	SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000017")),
						// 	AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
						// 	DrawerSeq: []model.Drawer{},
						// },
					},
					Canvas: model.Canvas{
						BoardName: "board1",
						AllArea:   25,
					},
				},
				members: []model.User{
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000002"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000003"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000004"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000005"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000006"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000007"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000008"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000009"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000010"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000011"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000012"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000013"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000014"))},
					{Id: model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000015"))},
					// model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000016")),
					// model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000017")),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupMemberRoles(tt.args.g, tt.args.members)
			t.Log(tt.args.g.Odais)
		})
	}
}
