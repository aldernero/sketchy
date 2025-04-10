![sketchy_logo](assets/images/logo.png)

Sketchy is a framework for making generative art in Go. It is inspired by [vsketch](https://github.com/abey79/vsketch) and [openFrameworks](https://github.com/openframeworks/openFrameworks). It uses [canvas](https://github.com/tdewolff/canvas) for drawing and the [ebiten](https://github.com/hajimehoshi/ebiten) game engine for the GUI. It's designed to provide controls (sliders, checkboxes, buttons) via simple JSON that can be used within a familiar `update()` and `draw()` framework to enable quick iteration on designs.

The [Getting Started](docs/getting-started.md) guide is a good place to start, and even walk through creating a "Hello Circle" sketch from scratch.

Below are a couple of screenshots from the example sketches:

### Fractal
![fractal_example](assets/images/fractal_example_screenshot.png)
### Noise

![noise_example](assets/images/opensimplex_example_screenshot.png)

### 10PRINT
![10print_example](assets/images/10print_example_screenshot.png)

# Installation

## Build Prerequisites
Sketchy requires Go version 1.23 or higher. It assumes that `go` is in the system path.

## Clone the repo

```shell
git clone https://github.com/aldernero/sketchy.git
```
## Compile the sketchy binary
```shell
go build -o sketchy ./cmd/sketchy/sketchy.go
```

## Running the examples
For any of the examples in the `examples` directory, run using standard go commands:
```shell
❯ cd ~/sketchy/examples/lissajous
❯ go run main.go
```

# Creating a new sketch

The syntax for creating a new sketch is `sketchy init project_name`. This will create a new directory with a configuration file and base sketch file:
```shell
❯ ./sketchy init mysketch
❯ tree mysketch
mysketch
├── go.mod
├── go.sum
├── icon.png
├── main.go
└── sketch.json
```
Sketchy init's a go module and runs `go mod tidy` to get all of the go dependencies.

The next step are to configure sketch parameter and controls in `sketch.json` and add the drawing code to `main.go`. See the `examples` directory and documentation for more details.

# Running a sketch

The syntax for running a sketch is `sketchy run project_name`. This is just a wrapper around running `go run main.go` from the project directory. Even the empty example above will run, althought you'll just see the 2 example controls and a blank drawing area.

# The Control Panel

The control panel contains both custom controls defined in the `sketch.json` file and builtin controls. Below is an example
![control_panel](assets/images/control_panel.png)

The builtins section controls the random seed and saving images and configurations. These also have keyboard shortcuts
that are listed in the next section of README. The control panel will not show up in saved images. If you close the 
control panel, you can reopen it by pressing the space bar.

# Saving sketches and configurations

There are three builtin keyboard shortcuts for saving sketch images and configurations:
- "s" key - saves the current frame as an SVG file. The filename has the format `<prefix>_<timestamp>.svg`, where `<prefix>` by default is the project name (what you used during `sketchy init project_name`)
- "p" key - same as above but saves the current frame as a PNG image.
- "c" key - saves the configuration (control values and sketch parameters) as JSON. The filename has the format `<prefix>_config_<timestamp>.json`, where `<prefix>` by default is the project name (what you used during `sketchy init project_name`)
- "Esc" key - saves a screenshot whole window, including the control panel. The filename has the format `screenshot_<timestamp>.png`
