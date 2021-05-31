package dockerfile2dot

import (
	"reflect"
	"testing"
)

func Test_dockerfileToSimplifiedDockerfile(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name string
		args args
		want SimplifiedDockerfile
	}{
		{
			name: "Most minimal Dockerfile",
			args: args{content: []byte(`
			FROM scratch
			`)},
			want: SimplifiedDockerfile{
				BaseImages: []BaseImage{
					{ID: "scratch"},
				},
				Stages: []Stage{
					{
						ID: "0", WaitFor: []WaitFor{
							{ID: "scratch", Type: waitForType(from)},
						},
					},
				},
			},
		},
		{
			name: "All waitFor types",
			args: args{content: []byte(`
			FROM ubuntu as base
			FROM scratch
			COPY --from=base . .
			RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build
			`)},
			want: SimplifiedDockerfile{
				BaseImages: []BaseImage{
					{ID: "ubuntu"},
					{ID: "scratch"},
					{ID: "buildcache"},
				},
				Stages: []Stage{
					{
						ID: "0", Name: "base", WaitFor: []WaitFor{
							{ID: "ubuntu", Type: waitForType(from)},
						},
					},
					{
						ID: "1", WaitFor: []WaitFor{
							{ID: "scratch", Type: waitForType(from)},
							{ID: "base", Type: waitForType(copy)},
							{ID: "buildcache", Type: waitForType(runMountTypeCache)},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dockerfileToSimplifiedDockerfile(tt.args.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dockerfileToSimplifiedDockerfile()\ngot  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}
