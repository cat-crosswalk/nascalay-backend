package random

import (
	"log"
	"testing"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/gofrs/uuid"
)

func TestSetupMemberRoles(t *testing.T) {
	type args struct {
		g       *model.Game
		members []model.UserId
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "setup roles",
			args: args{
				g: &model.Game{
					Odais: []model.Odai{
						{
							Title:     "odai1",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai2",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000002")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai3",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000003")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai4",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000004")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai5",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000005")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai6",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000006")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai7",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000007")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai8",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000008")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai9",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000009")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai10",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000010")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai11",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000011")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai12",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000012")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai13",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000013")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai14",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000014")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
						},
						{
							Title:     "odai15",
							SenderId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000015")),
							AnswerId:  model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")),
							DrawerSeq: []model.Drawer{},
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
				members: []model.UserId{
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000002")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000003")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000004")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000005")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000006")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000007")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000008")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000009")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000010")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000011")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000012")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000013")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000014")),
					model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000015")),
					// model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000016")),
					// model.UserId(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000017")),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupMemberRoles(tt.args.g, tt.args.members)
			log.Println(tt.args.g.Odais)
		})
	}
}
