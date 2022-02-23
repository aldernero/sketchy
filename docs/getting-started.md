This guide covers installation of sketchy and creating your first sketch.

# Installation

## Prerequisites
Sketchy requires Go version 1.17 or higher. It assumes that `go` is in the system path. If you are running Windows, install Windows Subsystem for Linux (WSL), so that you have `bash`, which is used by the install script.

## Clone the repo

```shell
git clone https://github.com/aldernero/sketchy.git
```
## Install sketchy environment
```shell
cd sketchy/scripts
./sketch_install.sh <target_directory>
```
This will create a directory `target_directory`, build the sketchy binary, and copy the binary and template files to the newly created directory.

Example:

```bash
❯ cd ~/sketchy/scripts
❯ ./sketchy_install.sh ~/sketchy_files
Sucessfully installed sketchy environment to /home/vernon/sketchy_files
❯ tree ~/sketchy_files
/home/vernon/sketchy_files
├── sketchy
└── template
    ├── main.go
    └── sketch.json

1 directory, 3 files
```
Sketchy is now installed and ready to run from `target_directory`.

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
├── main.go
└── sketch.json
```
Sketchy init's a go module and runs `go mod tidy` to get all of the go dependencies.

The next step are to configure sketch parameter and controls in `sketch.json` and add the drawing code to `main.go`. See the `examples` directory and documentation for more details.

# Example: creating a "Hello Circle" sketch
Rather than a typical "Hello World!" program, let's create something graphical that illustrates how to use controls and draw in the sketch area.

Create a new sketch called `hello_circle`
```shell
❯ ./sketchy init hello_circle
❯ cd hello_circle
❯ ls
go.mod  go.sum  main.go  sketch.json
```
So far this is identical to the previous section. Let's look at the contents of `sketch.json`:
```json
{
    "SketchWidth": 800,
    "SketchHeight": 800,
    "ControlWidth": 240,
    "Controls": [
        {
            "Name": "control1",
            "MinVal": 1,
            "MaxVal": 100,
            "Val": 10,
            "Incr": 1
        },
        {
            "Name": "control2",
            "MinVal": 0,
            "MaxVal": 2,
            "Val": 0.9,
            "Incr": 0.01
        }
    ]
}
```
This is the default configuration with 2 example controls. The first 3 lines define the sketch area size (800 x 800 pixels), and the the control area width (240 pixels). The "Controls" section lists the controls that will appear as sliders in the sketch. Let's make them more meaningful. The first one will represent the radius of a circle we draw in the sketch area. The second one will represent the line width of the circle. Change the values to the following:
```json
{
    "SketchWidth": 800,
    "SketchHeight": 800,
    "ControlWidth": 240,
    "Controls": [
        {
            "Name": "radius",
            "MinVal": 1,
            "MaxVal": 350,
            "Val": 200,
            "Incr": 1
        },
        {
            "Name": "thickness",
            "MinVal": 0.5,
            "MaxVal": 10,
            "Val": 1,
            "Incr": 0.5
        }
    ]
}
```
Notice that the radius can vary from 1 to 350 pixels and the line thickness can vary from 0.5 to 10 pixels.

Run the sketch to see the controls in action:
You can run `sketchy run hello_circle` from sketchy's base directory, or if you are inside the project directory, you can use go directly:
```shell
go run main.go
```
![Screenshot_20220222_212504](https://user-images.githubusercontent.com/96601789/155263278-221a2e98-f48e-4300-a07b-4c6c844d3aeb.png)

You should see 2 sliders in the control area on the left. You can change the values by clicking or dragging within the slider bar area. You can also use the mouse wheel to increment and decrement the value. The sketch area is blank at the moment, let's change that!

Close the sketch and open `main.go` in an editor. There are two functions `update` and `draw` where you implement the drawing. For a simple case like this we don't need `update`, we can do everything in the `draw` function.
```go
func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	// Drawing code goes here
}
```
Notice that the `draw` function takes two arguments. The first argument stores the Sketch struct used to store our sketch information, including the two slider controls. Here is how you get the value from a slider:

```go
val := s.Var("slider name")
```

The value will be a float64. For our case we could define two variables that are tied to the controls we defined earlier:
```go
radius := s.Var("radius")
thickness := s.Var("thickness")
```
Notice the argument to `Var` is the same name we used in `sketch.json`.

The other argument to `draw` is a `gg` drawing context. See the [gg](https://github.com/fogleman/gg) documentation for full details. For this example we will simply 1) set a drawing color, 2) set the line thickness, 3) define the circle object, and 4) draw the circle.  Here is the entire draw function:
```go
func draw(s *sketchy.Sketch, c *gg.Context) {
	// Drawing code goes here
	radius := s.Var("radius")
	thickness := s.Var("thickness")
	c.SetColor(color.White)
	c.SetLineWidth(thickness)
	c.DrawCircle(s.SketchWidth/2, s.SketchHeight/2, radius)
	c.Stroke()
}
```
The `DrawCircle` gg function takes 3 arguments: an x and y positions for the center of the circle and a radius. We can reference the `SketchWidth` and `SketchHeight` arguments directly on the sketch struct. Halving these values places the circle at the center of the drawing area.

Run the sketch again, and you should see a white circle in the sketch area, and you should be able to vary the radius and thickness with the sliders. 
![Screenshot_20220222_214157](https://user-images.githubusercontent.com/96601789/155263315-a90d8730-e049-4005-bf06-9bc5c0cc27f4.png)
Congratulations, you made your first sketch!

# Saving sketches and configurations

There are two builtin keyboard shortcuts for saving sketch images and configurations:
- "s" key - saves the current frame as a PNG. The filename has the format `<prefix>_<timestamp>.png`, where `<prefix>` by default is the project name (what you used during `sketchy init project_name`)
- "c" key - saves the configuration (control values and sketch parameters) as JSON. The filename has the format `<prefix>_config_<timestamp>.json`, where `<prefix>` by default is the project name (what you used during `sketchy init project_name`)
