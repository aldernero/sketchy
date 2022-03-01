This guide covers how to configure your sketch using the JSON configuration file, Go code, and command line parameters.

# The JSON Configuration File

Each sketch is accompanied by a JSON configuration file. Although everything could be configured in code, having a JSON format speeds up development, making it especially easy to configure the controls that will appear in your sketch. Below is the minimal configuration file that's created when running `sketchy init project_name`:
```json
{
  "SketchWidth": 800,
  "SketchHeight": 800,
  "ControlWidth": 240,
  "Sliders": [
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
  ],
  "Toggles": [
    {
      "Name": "checkbox",
      "Checked": false
    },
    {
      "Name": "button",
      "IsButton": true
    }
  ],
  "SketchBackgroundColor": "#1e1e1e",
  "SketchOutlineColor": "#ffdb00",
  "ControlBackgroundColor": "#1e1e1e",
  "ControlOutlineColor": "#ffdb00"
}
```
The first three parameters setup the window size, most importantly the size of the sketch area, and also the width of
the control area, where the controls will appear. In the `Sliders` section you define the sliders for your sketch. In 
the `Toggles` section you define the checkboxes and buttons for your sketch. In code you can reference controls and tie 
them to a variable. For each slider you define the range of values it supports along with an initial value and the step
size `Incr` between values. The result is always a `float64` in code. For each toggle you specify if it's a button
or a checkbox (default), and whether the initial state is checked.

There are other parameters not listed in the template. Here are the missing parameters for completeness:

## Sketch Parameters
| Parameter                 | Type   | Default   | Description                            |
|---------------------------|--------|-----------|----------------------------------------|
| Title                     | String | "Sketch"  | Sketch title                           |
| Prefix                    | string | "sketch"  | prefix for filenames                   |
| SketchWidth               | float  | 800       | sketch area width in pixels            |
| SketchHeight              | float  | 800       | sketch area height in pixels           |
| ControlWidth              | float  | 240       | control area width in pixels           |
| RandomSeed                | int    | 0         | seed to builtin PRNG                   |
| SketchBackgroundColor     | string | "#1e1e1e" | sketch area background color           |
| SketchOutlineColor        | string | "#ffdb00" | sketch area outline color              |
| ControlBackgroundColor    | string | "#1e1e1e" | control area background color          |
| ControlOutlineColor       | string | "#ffdb00" | control area background color          |
| DisableClearBetweenFrames | bool   | false     | don't clear sketch area between frames |

## Slider Parameters
| Parameter       | Type   | Default      | Description              |
|-----------------|--------|--------------|--------------------------|
| Width           | float  | ControlWidth | control width            |
| Height          | float  | 15px         | control height           |
| BackgroundColor | string | "#1e1e1e"    | control background color |
| OutlineColor    | string | "#ffdb00"    | control outline color    |
| FillColor       | string | "#ffdb00"    | control fill color       |
| TextColor       | string | "#ffffff"    | control text color       |
| MinVal          | float  | 0            | minimum control value    |
| MaxVal          | float  | 0            | maximum control value    |
| Val             | float  | 0            | initial control value    | 
| Incr            | float  | 0            | increment value          |

## Toggle Parameters
| Parameter       | Type   | Default      | Description              |
|-----------------|--------|--------------|--------------------------|
| Width           | float  | ControlWidth | control width            |
| Height          | float  | 15px         | control height           |
| BackgroundColor | string | "#1e1e1e"    | control background color |
| OutlineColor    | string | "#ffdb00"    | control outline color    |
| FillColor       | string | "#ffdb00"    | control fill color       |
| TextColor       | string | "#ffffff"    | control text color       |
 | Checked         | bool   | false        | initial control state    |
| IsButton        | bool   | false        | is the control a button  |
# Referencing Sketch Parameters in Code

The Sketch parameters can be referenced like `s.<ParameterName>` in the `update` and `draw` functions. The value of each control can be referenced via `s.Slider("<control name>")`, where `<control name>` is the name you used in the JSON configuration file.


# Command Line Parameters

There are 3 command line arguments that may be useful for setting things at runtime. They are
- `-c <config file>` : The configuration file to use (default: "sketch.json")
- `-p <prefix>` : The filename prefix to use when saving sketch images and configuration files
- `-s <seed>` : The seed for the builtin random number generator