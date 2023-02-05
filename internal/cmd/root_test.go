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
  -c, --concentrate             concentrate the edges (default false)
  -d, --dpi uint                dots per inch of the PNG export (default 96)
  -e, --edgestyle               style of the graph edges, one of: default, solid (default default)
  -f, --filename string         name of the Dockerfile (default "Dockerfile")
  -h, --help                    help for dockerfilegraph
      --layers                  display all layers (default false)
      --legend                  add a legend (default false)
  -m, --max-label-length uint   maximum length of the node labels, must be at least 4 (default 20)
  -n, --nodesep float           minimum space between two adjacent nodes in the same rank (default 1)
  -o, --output                  output file format, one of: canon, dot, pdf, png, raw, svg (default pdf)
  -r, --ranksep float           minimum separation between ranks (default 0.5)
  -u, --unflatten uint          stagger length of leaf edges between [1,u] (default 0)
      --version                 display the version of dockerfilegraph
`

var dockerfileContent = `
### TLS root certs and non-root user
FROM ubuntu:latest AS ubuntu

RUN \
  apt-get update \
  && apt-get install -y --no-install-recommends \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# ---

FROM golang:1.19 AS build-tool-dependencies
RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build

# ---

FROM scratch AS release

COPY --from=ubuntu /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-tool-dependencies . .

ENTRYPOINT ["/example"]
`

func TestRootCmd(t *testing.T) {
	tests := []test{
		{
			name:    "help flag",
			cliArgs: []string{"--help"},
			wantOut: `dockerfilegraph visualizes your multi-stage Dockerfile.
It creates a visual graph representation of the build process.

` + usage,
		},
		{
			name:    "version flag",
			cliArgs: []string{"--version"},
			wantOut: `{` +
				`"GitVersion":"v0.0.0-dev",` +
				`"GitCommit":"da39a3ee5e6b4b0d3255bfef95601890afd80709",` +
				`"BuildDate":"0000-00-00T00:00:00Z"}
`,
		},
		{
			name:    "no args",
			wantOut: "Successfully created Dockerfile.pdf\n",
		},
		{
			name:              "empty Dockerfile",
			dockerfileContent: " ", // space is needed so that the default Dockerfile is not used
			wantErr:           true,
			wantOut:           "Error: file with no instructions\n" + usage + "\n",
		},
		{
			name:    "--max-label-length too small",
			cliArgs: []string{"--max-label-length", "3"},
			wantErr: true,
			wantOut: "Error: --max-label-length must be at least 4\n" + usage + "\n",
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
			name:        "output flag svg",
			cliArgs:     []string{"--output", "svg"},
			wantOut:     "Successfully created Dockerfile.svg\n",
			wantOutFile: "Dockerfile.svg",
		},
		{
			name:        "filename flag",
			cliArgs:     []string{"--filename", "subdir/../Dockerfile"},
			wantOut:     "Successfully created Dockerfile.pdf\n",
			wantOutFile: "Dockerfile.pdf",
		},
		{
			name:         "filename flag with missing Dockerfile",
			cliArgs:      []string{"--filename", "Dockerfile.missing"},
			wantErr:      true,
			wantOutRegex: "^Error: could not find a Dockerfile at .+Dockerfile.missing\n",
		},
		{
			name:        "layers flag",
			cliArgs:     []string{"--layers", "-o", "raw"},
			wantOut:     "Successfully created Dockerfile.raw\n",
			wantOutFile: "Dockerfile.raw",
			//nolint:lll
			wantOutFileContent: `digraph G {
	compound=true;
	nodesep=1.00;
	rankdir=LR;
	ranksep=0.50;
	stage_0_layer_0->stage_0_layer_1;
	external_image_0->stage_0_layer_0;
	stage_1_layer_0->stage_1_layer_1;
	external_image_1->stage_1_layer_0;
	external_image_2->stage_1_layer_1[ arrowhead=ediamond, style=dotted ];
	stage_2_layer_0->stage_2_layer_1;
	stage_2_layer_1->stage_2_layer_2;
	stage_2_layer_2->stage_2_layer_3;
	external_image_3->stage_2_layer_0;
	stage_0_layer_1->stage_2_layer_1[ arrowhead=empty, ltail=cluster_stage_0, style=dashed ];
	stage_1_layer_1->stage_2_layer_2[ arrowhead=empty, ltail=cluster_stage_1, style=dashed ];
	subgraph cluster_stage_0 {
	label=ubuntu;
	margin=16;
	stage_0_layer_0 [ fillcolor=white, label="FROM ubuntu:lates...", penwidth=0.5, shape=box, style="filled,rounded", width=2 ];
	stage_0_layer_1 [ fillcolor=white, label="RUN   apt-get upd...", penwidth=0.5, shape=box, style="filled,rounded", width=2 ];

}
;
	subgraph cluster_stage_1 {
	label="build-tool-dependencies";
	margin=16;
	stage_1_layer_0 [ fillcolor=white, label="FROM golang:1.19 ...", penwidth=0.5, shape=box, style="filled,rounded", width=2 ];
	stage_1_layer_1 [ fillcolor=white, label="RUN --mount=type=...", penwidth=0.5, shape=box, style="filled,rounded", width=2 ];

}
;
	subgraph cluster_stage_2 {
	fillcolor=grey90;
	label=release;
	margin=16;
	style=filled;
	stage_2_layer_0 [ fillcolor=white, label="FROM scratch AS r...", penwidth=0.5, shape=box, style="filled,rounded", width=2 ];
	stage_2_layer_1 [ fillcolor=white, label="COPY --from=ubunt...", penwidth=0.5, shape=box, style="filled,rounded", width=2 ];
	stage_2_layer_2 [ fillcolor=white, label="COPY --from=build...", penwidth=0.5, shape=box, style="filled,rounded", width=2 ];
	stage_2_layer_3 [ fillcolor=white, label="ENTRYPOINT ['/exa...", penwidth=0.5, shape=box, style="filled,rounded", width=2 ];

}
;
	external_image_0 [ color=grey20, fontcolor=grey20, label="ubuntu:latest", shape=box, style="dashed,rounded", width=2 ];
	external_image_1 [ color=grey20, fontcolor=grey20, label="golang:1.19", shape=box, style="dashed,rounded", width=2 ];
	external_image_2 [ color=grey20, fontcolor=grey20, label="buildcache", shape=box, style="dashed,rounded", width=2 ];
	external_image_3 [ color=grey20, fontcolor=grey20, label="scratch", shape=box, style="dashed,rounded", width=2 ];

}
`,
		},
		{
			name:        "layers flag with solid edges",
			cliArgs:     []string{"--layers", "-e", "solid", "-o", "canon"},
			wantOut:     "Successfully created Dockerfile.canon\n",
			wantOutFile: "Dockerfile.canon",
			wantOutFileContent: `digraph G {
	graph [compound=true,
		nodesep=1.00,
		rankdir=LR,
		ranksep=0.50
	];
	node [label="\N"];
	subgraph cluster_stage_0 {
		graph [label=ubuntu,
			margin=16
		];
		stage_0_layer_0	[fillcolor=white,
			label="FROM ubuntu:lates...",
			penwidth=0.5,
			shape=box,
			style="filled,rounded",
			width=2];
		stage_0_layer_1	[fillcolor=white,
			label="RUN   apt-get upd...",
			penwidth=0.5,
			shape=box,
			style="filled,rounded",
			width=2];
		stage_0_layer_0 -> stage_0_layer_1;
	}
	subgraph cluster_stage_1 {
		graph [label="build-tool-dependencies",
			margin=16
		];
		stage_1_layer_0	[fillcolor=white,
			label="FROM golang:1.19 ...",
			penwidth=0.5,
			shape=box,
			style="filled,rounded",
			width=2];
		stage_1_layer_1	[fillcolor=white,
			label="RUN --mount=type=...",
			penwidth=0.5,
			shape=box,
			style="filled,rounded",
			width=2];
		stage_1_layer_0 -> stage_1_layer_1;
	}
	subgraph cluster_stage_2 {
		graph [fillcolor=grey90,
			label=release,
			margin=16,
			style=filled
		];
		stage_2_layer_0	[fillcolor=white,
			label="FROM scratch AS r...",
			penwidth=0.5,
			shape=box,
			style="filled,rounded",
			width=2];
		stage_2_layer_1	[fillcolor=white,
			label="COPY --from=ubunt...",
			penwidth=0.5,
			shape=box,
			style="filled,rounded",
			width=2];
		stage_2_layer_0 -> stage_2_layer_1;
		stage_2_layer_2	[fillcolor=white,
			label="COPY --from=build...",
			penwidth=0.5,
			shape=box,
			style="filled,rounded",
			width=2];
		stage_2_layer_1 -> stage_2_layer_2;
		stage_2_layer_3	[fillcolor=white,
			label="ENTRYPOINT ['/exa...",
			penwidth=0.5,
			shape=box,
			style="filled,rounded",
			width=2];
		stage_2_layer_2 -> stage_2_layer_3;
	}
	stage_0_layer_1 -> stage_2_layer_1	[arrowhead=empty,
		ltail=cluster_stage_0];
	external_image_0	[color=grey20,
		fontcolor=grey20,
		label="ubuntu:latest",
		shape=box,
		style="dashed,rounded",
		width=2];
	external_image_0 -> stage_0_layer_0;
	stage_1_layer_1 -> stage_2_layer_2	[arrowhead=empty,
		ltail=cluster_stage_1];
	external_image_1	[color=grey20,
		fontcolor=grey20,
		label="golang:1.19",
		shape=box,
		style="dashed,rounded",
		width=2];
	external_image_1 -> stage_1_layer_0;
	external_image_2	[color=grey20,
		fontcolor=grey20,
		label=buildcache,
		shape=box,
		style="dashed,rounded",
		width=2];
	external_image_2 -> stage_1_layer_1	[arrowhead=ediamond];
	external_image_3	[color=grey20,
		fontcolor=grey20,
		label=scratch,
		shape=box,
		style="dashed,rounded",
		width=2];
	external_image_3 -> stage_2_layer_0;
}
`,
		},
		{
			name:        "legend flag with concentrated edges and unflattened",
			cliArgs:     []string{"--legend", "-c", "-u", "2", "-o", "canon"},
			wantOut:     "Successfully created Dockerfile.canon\n",
			wantOutFile: "Dockerfile.canon",
			wantOutFileContent: `digraph G {
	graph [compound=true,
		concentrate=true,
		nodesep=1.00,
		rankdir=LR,
		ranksep=0.50
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
		key:i1:e -> key2:i1:w	[arrowhead=empty,
			style=dashed];
		key:i2:e -> key2:i2:w	[arrowhead=ediamond,
			style=dotted];
	}
	external_image_0	[color=grey20,
		fontcolor=grey20,
		label="ubuntu:latest",
		shape=box,
		style="dashed,rounded",
		width=2];
	stage_0	[label=ubuntu,
		shape=box,
		style=rounded,
		width=2];
	external_image_0 -> stage_0	[minlen=1];
	stage_2	[fillcolor=grey90,
		label=release,
		shape=box,
		style="filled,rounded",
		width=2];
	stage_0 -> stage_2	[arrowhead=empty,
		style=dashed];
	external_image_1	[color=grey20,
		fontcolor=grey20,
		label="golang:1.19",
		shape=box,
		style="dashed,rounded",
		width=2];
	stage_1	[label="build-tool-depend...",
		shape=box,
		style=rounded,
		width=2];
	external_image_1 -> stage_1	[minlen=1];
	stage_1 -> stage_2	[arrowhead=empty,
		style=dashed];
	external_image_2	[color=grey20,
		fontcolor=grey20,
		label=buildcache,
		shape=box,
		style="dashed,rounded",
		width=2];
	external_image_2 -> stage_1	[arrowhead=ediamond,
		minlen=2,
		style=dotted];
	external_image_3	[color=grey20,
		fontcolor=grey20,
		label=scratch,
		shape=box,
		style="dashed,rounded",
		width=2];
	external_image_3 -> stage_2	[minlen=1];
}
`,
		},
		{
			name:        "legend flag with solid edges",
			cliArgs:     []string{"--legend", "-e", "solid", "-o", "canon"},
			wantOut:     "Successfully created Dockerfile.canon\n",
			wantOutFile: "Dockerfile.canon",
			wantOutFileContent: `digraph G {
	graph [compound=true,
		nodesep=1.00,
		rankdir=LR,
		ranksep=0.50
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
	external_image_0	[color=grey20,
		fontcolor=grey20,
		label="ubuntu:latest",
		shape=box,
		style="dashed,rounded",
		width=2];
	stage_0	[label=ubuntu,
		shape=box,
		style=rounded,
		width=2];
	external_image_0 -> stage_0;
	stage_2	[fillcolor=grey90,
		label=release,
		shape=box,
		style="filled,rounded",
		width=2];
	stage_0 -> stage_2	[arrowhead=empty];
	external_image_1	[color=grey20,
		fontcolor=grey20,
		label="golang:1.19",
		shape=box,
		style="dashed,rounded",
		width=2];
	stage_1	[label="build-tool-depend...",
		shape=box,
		style=rounded,
		width=2];
	external_image_1 -> stage_1;
	stage_1 -> stage_2	[arrowhead=empty];
	external_image_2	[color=grey20,
		fontcolor=grey20,
		label=buildcache,
		shape=box,
		style="dashed,rounded",
		width=2];
	external_image_2 -> stage_1	[arrowhead=ediamond];
	external_image_3	[color=grey20,
		fontcolor=grey20,
		label=scratch,
		shape=box,
		style="dashed,rounded",
		width=2];
	external_image_3 -> stage_2;
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
				"Output mismatch (-want +got):\n%s",
				cmp.Diff(tt.wantOutRegex, buf.String()),
			)
		}
	}
}
