package config

import (
	"testing"
	"time"
)

func TestValue_AddHistory(t *testing.T) {
	now := time.Now()
	type fields struct {
		History []HistoryEntry
	}
	type args struct {
		uuid      string
		station   string
		song      string
		timestamp time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []HistoryEntry
	}{
		{
			name: "new station, new song, should be added",
			fields: fields{
				History: []HistoryEntry{
					{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-120 * time.Second)},
					{Uuid: "1", Station: "station1", Song: "song12", Timestamp: now.Add(-121 * time.Second)},
					{Uuid: "2", Station: "station2", Song: "song21", Timestamp: now.Add(-122 * time.Second)},
				},
			},
			args: args{uuid: "3", station: "station3", song: "song31", timestamp: now},
			want: []HistoryEntry{
				{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-120 * time.Second)},
				{Uuid: "1", Station: "station1", Song: "song12", Timestamp: now.Add(-121 * time.Second)},
				{Uuid: "2", Station: "station2", Song: "song21", Timestamp: now.Add(-122 * time.Second)},
				{Uuid: "3", Station: "station3", Song: "song31", Timestamp: now},
			},
		},
		{
			name: "same station, same song, recently played, should be skipped",
			fields: fields{
				History: []HistoryEntry{
					{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-120 * time.Second)},
					{Uuid: "1", Station: "station1", Song: "song12", Timestamp: now.Add(-40 * time.Second)},
					{Uuid: "2", Station: "station2", Song: "song21", Timestamp: now.Add(-30 * time.Second)},
				},
			},
			args: args{uuid: "1", station: "station1", song: "song12", timestamp: now},
			want: []HistoryEntry{
				{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-120 * time.Second)},
				{Uuid: "1", Station: "station1", Song: "song12", Timestamp: now.Add(-40 * time.Second)},
				{Uuid: "2", Station: "station2", Song: "song21", Timestamp: now.Add(-30 * time.Second)},
			},
		},
		{
			name: "same station, new non-empty song, should add",
			fields: fields{
				History: []HistoryEntry{
					{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-120 * time.Second)},
					{Uuid: "1", Station: "station1", Song: "song12", Timestamp: now.Add(-121 * time.Second)},
					{Uuid: "2", Station: "station2", Song: "song21", Timestamp: now.Add(-122 * time.Second)},
				},
			},
			args: args{uuid: "1", station: "station1", song: "song13", timestamp: now},
			want: []HistoryEntry{
				{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-120 * time.Second)},
				{Uuid: "1", Station: "station1", Song: "song12", Timestamp: now.Add(-121 * time.Second)},
				{Uuid: "2", Station: "station2", Song: "song21", Timestamp: now.Add(-122 * time.Second)},
				{Uuid: "1", Station: "station1", Song: "song13", Timestamp: now},
			},
		},
		{
			name: "same station, same song, not recently played, not last played, should add",
			fields: fields{
				History: []HistoryEntry{
					{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-16 * time.Minute)},
					{Uuid: "1", Station: "station1", Song: "song12", Timestamp: now.Add(-121 * time.Second)},
					{Uuid: "2", Station: "station2", Song: "song21", Timestamp: now.Add(-122 * time.Second)},
				},
			},
			args: args{uuid: "1", station: "station1", song: "song11", timestamp: now},
			want: []HistoryEntry{
				{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-16 * time.Minute)},
				{Uuid: "1", Station: "station1", Song: "song12", Timestamp: now.Add(-121 * time.Second)},
				{Uuid: "2", Station: "station2", Song: "song21", Timestamp: now.Add(-122 * time.Second)},
				{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now},
			},
		},
		{
			name: "same station, same song, not recently played, but last played, should skip",
			fields: fields{
				History: []HistoryEntry{
					{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-16 * time.Minute)},
				},
			},
			args: args{uuid: "1", station: "station1", song: "song11", timestamp: now},
			want: []HistoryEntry{
				{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now.Add(-16 * time.Minute)},
			},
		},
		{
			name: "same station, new song after empty song, should replace",
			fields: fields{
				History: []HistoryEntry{
					{Uuid: "2", Station: "station2", Song: "song2", Timestamp: now.Add(-26 * time.Second)},
					{Uuid: "1", Station: "station1", Song: "", Timestamp: now.Add(-16 * time.Second)},
				},
			},
			args: args{uuid: "1", station: "station1", song: "song11", timestamp: now},
			want: []HistoryEntry{
				{Uuid: "2", Station: "station2", Song: "song2", Timestamp: now.Add(-26 * time.Second)},
				{Uuid: "1", Station: "station1", Song: "song11", Timestamp: now},
			},
		},
		{
			name: "same station, new empty song after non-empty song, should skip",
			fields: fields{
				History: []HistoryEntry{
					{Uuid: "1", Station: "station1", Song: "", Timestamp: now.Add(-16 * time.Second)},
					{Uuid: "2", Station: "station2", Song: "song2", Timestamp: now.Add(-6 * time.Second)},
				},
			},
			args: args{uuid: "2", station: "station2", song: "", timestamp: now},
			want: []HistoryEntry{
				{Uuid: "1", Station: "station1", Song: "", Timestamp: now.Add(-16 * time.Second)},
				{Uuid: "2", Station: "station2", Song: "song2", Timestamp: now.Add(-6 * time.Second)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				History: tt.fields.History,
			}
			v.upsertHistory(tt.args.timestamp, tt.args.uuid, tt.args.station, tt.args.song)
			if len(v.History) != len(tt.want) {
				t.Errorf("test=%q got history length=%v, want=%v", tt.name, len(v.History), len(tt.want))
			}
			for i := 0; i < len(tt.want); i++ {
				wantIx := tt.want[i]
				gotIx := v.History[i]
				if wantIx.Uuid != gotIx.Uuid {
					t.Errorf("test=%q ix=%d: got uuid=%v, want uuid=%v",
						tt.name, i, gotIx.Uuid, wantIx.Uuid)
				}
				if wantIx.Station != gotIx.Station {
					t.Errorf("test=%q ix=%d: got station name=%v, want station name=%v",
						tt.name, i, gotIx.Station, wantIx.Station)
				}
				if wantIx.Song != gotIx.Song {
					t.Errorf("test=%q ix=%d: got Song=%v, want Song=%v",
						tt.name, i, gotIx.Song, wantIx.Song)
				}
				if wantIx.Timestamp != gotIx.Timestamp {
					t.Errorf("test=%q ix=%d: got Timestamp=%v, want Timestamp=%v",
						tt.name, i, gotIx.Timestamp, wantIx.Timestamp)
				}
			}
		})
	}
}
