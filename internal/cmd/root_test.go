package cmd_test

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/patrickhoefler/dockerfilegraph/internal/cmd"
	"github.com/spf13/afero"
)

type test struct {
	name               string
	cliArgs            []string
	dockerfileContent  string
	wantErr            bool
	wantOut            string
	wantOutRegex       string
	wantOutFile        string
	wantOutFileContent string
}

var usage = `Usage:
  dockerfilegraph [flags]

Flags:
  -d, --dpi int   dots per inch of the PNG export (default 96)
  -h, --help      help for dockerfilegraph
      --layers    display all layers (default false)
      --legend    add a legend (default false)
  -o, --output    output file format, one of: canon, dot, pdf, png (default pdf)
`

// Taken from example/Dockerfile.
var dockerfileContent = `
### TLS root certs and non-root user
FROM ubuntu:latest AS ubuntu

RUN \
  # Note that the lack of a "lock" mechanism for apt dependencies
  # currently prevents a fully reproducible build
  apt-get update \
  && apt-get install -y --no-install-recommends \
  # Install TLS root certificates
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# ---

FROM golang:1.18 AS build
RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build

# ---

### Release image
FROM scratch AS release

# Copy the TLS certificates for encrypted network communication
COPY --from=ubuntu /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build . .

ENTRYPOINT ["/example"]
`

func TestRootCmd(t *testing.T) {
	tests := []test{
		{
			name:    "help flag",
			cliArgs: []string{"--help"},
			wantOut: `dockerfilegraph visualizes your multi-stage Dockerfile.
It outputs a graph representation of the build process.

` + usage,
		},
		{
			name:    "no args",
			wantOut: "Successfully created Dockerfile.pdf\n",
		},
		{
			name:              "empty Dockerfile",
			dockerfileContent: " ", // space is needed so that the default Dockerfile is not used
			wantErr:           true,
			wantOut: `Error: file with no instructions
` + usage + "\n",
		},
		{
			name:        "output flag canon",
			cliArgs:     []string{"--output", "canon"},
			wantOut:     "Successfully created Dockerfile.canon\n",
			wantOutFile: "Dockerfile.canon",
			wantOutFileContent: `digraph G {
	graph [nodesep=1,
		rankdir=LR
	];
	node [label="\N"];
	"ubuntu:latest"	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	0	[label=ubuntu,
		shape=Mrecord,
		width=2];
	"ubuntu:latest" -> 0;
	2	[fillcolor=grey90,
		label=release,
		shape=Mrecord,
		style=filled,
		width=2];
	0 -> 2	[arrowhead=empty];
	"golang:1.18"	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	1	[label=build,
		shape=Mrecord,
		width=2];
	"golang:1.18" -> 1;
	1 -> 2	[arrowhead=empty];
	buildcache	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	buildcache -> 1	[arrowhead=ediamond];
	scratch	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	scratch -> 2;
}
`,
		},
		{
			name:        "output flag dot",
			cliArgs:     []string{"--output", "dot"},
			wantOut:     "Successfully created Dockerfile.dot\n",
			wantOutFile: "Dockerfile.dot",
		},
		{
			name:        "output flag pdf",
			cliArgs:     []string{"-o", "pdf"},
			wantOut:     "Successfully created Dockerfile.pdf\n",
			wantOutFile: "Dockerfile.pdf",
		},
		{
			name:        "output flag png",
			cliArgs:     []string{"--output", "png"},
			wantOut:     "Successfully created Dockerfile.png\n",
			wantOutFile: "Dockerfile.png",
		},
		{
			name:        "output flag png with dpi",
			cliArgs:     []string{"--output", "png", "--dpi", "200"},
			wantOut:     "Successfully created Dockerfile.png\n",
			wantOutFile: "Dockerfile.png",
		},
		{
			name:        "layers flag",
			cliArgs:     []string{"--layers", "-o", "canon"},
			wantOut:     "Successfully created Dockerfile.canon\n",
			wantOutFile: "Dockerfile.canon",
			wantOutFileContent: `digraph G {
	graph [nodesep=0.03,
		rankdir=LR
	];
	node [label="\N"];
	subgraph cluster_0 {
		graph [label=ubuntu,
			style=rounded
		];
		0	[label=ubuntu,
			shape=Mrecord,
			style=invis,
			width=2];
		stage_0_layer_0	[label="FROM...",
			shape=Mrecord,
			style=dashed,
			width=2];
		stage_0_layer_1	[label="RUN...",
			shape=Mrecord,
			style=dashed,
			width=2];
	}
	subgraph cluster_1 {
		graph [label=build,
			style=rounded
		];
		1	[label=build,
			shape=Mrecord,
			style=invis,
			width=2];
		stage_1_layer_0	[label="FROM...",
			shape=Mrecord,
			style=dashed,
			width=2];
		stage_1_layer_1	[label="RUN...",
			shape=Mrecord,
			style=dashed,
			width=2];
	}
	subgraph cluster_2 {
		graph [label=release,
			style=rounded
		];
		2	[fillcolor=grey90,
			label=release,
			shape=Mrecord,
			style=invis,
			width=2];
		stage_2_layer_0	[fillcolor=grey90,
			label="FROM...",
			shape=Mrecord,
			style=dashed,
			width=2];
		stage_2_layer_1	[fillcolor=grey90,
			label="COPY...",
			shape=Mrecord,
			style=dashed,
			width=2];
		stage_2_layer_2	[fillcolor=grey90,
			label="COPY...",
			shape=Mrecord,
			style=dashed,
			width=2];
		stage_2_layer_3	[fillcolor=grey90,
			label="ENTRYPOINT...",
			shape=Mrecord,
			style=dashed,
			width=2];
	}
	"ubuntu:latest"	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	"ubuntu:latest" -> 0;
	0 -> 2	[arrowhead=empty];
	"golang:1.18"	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	"golang:1.18" -> 1;
	1 -> 2	[arrowhead=empty];
	buildcache	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	buildcache -> 1	[arrowhead=ediamond];
	scratch	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	scratch -> 2;
}
`,
		},
		{
			name:        "legend flag",
			cliArgs:     []string{"--legend", "-o", "canon"},
			wantOut:     "Successfully created Dockerfile.canon\n",
			wantOutFile: "Dockerfile.canon",
			wantOutFileContent: `digraph G {
	graph [nodesep=1,
		rankdir=LR
	];
	node [label="\N"];
	subgraph cluster_legend {
		key	[fontname=monospace,
			fontsize=10,
			label=<<table border="0" cellpadding="2" cellspacing="0" cellborder="0">
	<tr><td align="right" port="i0">FROM&nbsp;...&nbsp;</td></tr>
	<tr><td align="right" port="i1">COPY --from=...&nbsp;</td></tr>
	<tr><td align="right" port="i2">RUN --mount=type=cache,from=...&nbsp;</td></tr>
</table>>,
			shape=plaintext];
		key2	[fontname=monospace,
			fontsize=10,
			label=<<table border="0" cellpadding="2" cellspacing="0" cellborder="0">
	<tr><td port="i0">&nbsp;</td></tr>
	<tr><td port="i1">&nbsp;</td></tr>
	<tr><td port="i2">&nbsp;</td></tr>
</table>>,
			shape=plaintext];
		key:i0:e -> key2:i0:w;
		key:i1:e -> key2:i1:w	[arrowhead=empty];
		key:i2:e -> key2:i2:w	[arrowhead=ediamond];
	}
	"ubuntu:latest"	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	0	[label=ubuntu,
		shape=Mrecord,
		width=2];
	"ubuntu:latest" -> 0;
	2	[fillcolor=grey90,
		label=release,
		shape=Mrecord,
		style=filled,
		width=2];
	0 -> 2	[arrowhead=empty];
	"golang:1.18"	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	1	[label=build,
		shape=Mrecord,
		width=2];
	"golang:1.18" -> 1;
	1 -> 2	[arrowhead=empty];
	buildcache	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	buildcache -> 1	[arrowhead=ediamond];
	scratch	[color=grey20,
		fontcolor=grey20,
		shape=Mrecord,
		style=dashed,
		width=2];
	scratch -> 2;
}
`,
		},
	}

	for _, tt := range tests {
		// Create a fake filesystem for the input Dockerfile
		inputFS := afero.NewMemMapFs()
		if tt.dockerfileContent == "" {
			tt.dockerfileContent = dockerfileContent
		}
		_ = afero.WriteFile(inputFS, "Dockerfile", []byte(tt.dockerfileContent), 0644)

		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			command := cmd.NewRootCmd(buf, inputFS)
			command.SetArgs(tt.cliArgs)

			// Redirect Cobra output
			command.SetOut(buf)
			command.SetErr(buf)

			err := command.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: Execute() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			checkWantOut(t, tt, buf)

			if tt.wantOutFile != "" {
				_, err := os.Stat(tt.wantOutFile)
				if err != nil {
					t.Errorf("%s: %v", tt.name, err)
				}
			}

			if tt.wantOutFileContent != "" {
				outFileContent, err := os.ReadFile(tt.wantOutFile)
				if err != nil {
					t.Errorf("%s: %v", tt.name, err)
				}
				if diff := cmp.Diff(tt.wantOutFileContent, string(outFileContent)); diff != "" {
					t.Errorf("Output mismatch (-want +got):\n%s", diff)
				}
			}
		})

		// Cleanup
		if tt.wantOutFile != "" {
			os.Remove(tt.wantOutFile)
		}
	}
}

func TestExecute(t *testing.T) {
	tests := []test{
		{
			name:        "should work",
			wantOutFile: "Dockerfile.pdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.WriteFile("Dockerfile", []byte(dockerfileContent), 0644)

			cmd.Execute()

			if tt.wantOutFile != "" {
				_, err := os.Stat(tt.wantOutFile)
				if err != nil {
					t.Errorf("%s: %v", tt.name, err)
				}
			}

			// Cleanup
			os.Remove("Dockerfile")
			os.Remove(tt.wantOutFile)
		})
	}
}

func checkWantOut(t *testing.T, tt test, buf *bytes.Buffer) {
	if tt.wantOut == "" && tt.wantOutRegex == "" {
		t.Fatalf("Either wantOut or wantOutRegex must be set")
	}
	if tt.wantOut != "" && tt.wantOutRegex != "" {
		t.Fatalf("wantOut and wantOutRegex cannot be set at the same time")
	}

	if tt.wantOut != "" {
		if diff := cmp.Diff(tt.wantOut, buf.String()); diff != "" {
			t.Errorf("Output mismatch (-want +got):\n%s", diff)
		}
	} else if tt.wantOutRegex != "" {
		matched, err := regexp.Match(tt.wantOutRegex, buf.Bytes())
		if err != nil {
			t.Errorf("Error compiling regex: %v", err)
		}
		if !matched {
			t.Errorf(
				"Error matching regex: %v, output: %s",
				tt.wantOutRegex, buf.String(),
			)
		}
	}
}
