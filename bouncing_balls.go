package main

import (
  "github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
  "github.com/nullboundary/glfont"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
  mgl "github.com/go-gl/mathgl/mgl32"
  "strings"
  "bufio"
	"log"
	"math"
	"math/rand"
	"os"
  "fmt"
	"runtime"
)

var (
	ballRadius = 2
	ballMass   = 0.2

	space       *chipmunk.Space
	balls       []*chipmunk.Shape
	staticLines []*chipmunk.Shape
  staticBody  *chipmunk.Body
  staticBody2  *chipmunk.Body
	deg2rad     = math.Pi / 180
  Inf = vect.Float(math.Inf(1))
  count int
  font *glfont.Font

  vertAttrib uint32
)

const windowHeight = 800
const windowWidth = 1200

// drawCircle draws a circle for the specified radius, rotation angle, and the specified number of sides
func drawCircle(radius float64, sides int32, x float32, y float32) {
}

func addBall() {
	x := rand.Intn(50) + 200
  rad := float64(3 * rand.Float32() + 2)
  mass := float32(math.Pow(rad, 2) * math.Pi * 0.8)
	ball := chipmunk.NewCircle(vect.Vector_Zero, float32(rad))
	ball.SetElasticity(0.6)
  ball.SetFriction(0.9)

	body := chipmunk.NewBody(vect.Float(mass), ball.Moment(float32(mass)))
	body.SetPosition(vect.Vect{vect.Float(x), windowHeight})
	body.SetAngle(vect.Float(rand.Float32() * 2 * math.Pi))

	body.AddShape(ball)
	space.AddBody(body)
	balls = append(balls, ball)
}

func addBigBall() {
	x := rand.Intn(150) + 300
  rad := float64(80 * rand.Float32() + 5)
  mass := float32(math.Pow(rad, 2) * math.Pi * 0.3)
	ball := chipmunk.NewCircle(vect.Vector_Zero, float32(rad))
	ball.SetElasticity(0.7)
  ball.SetFriction(0.9)

	body := chipmunk.NewBody(vect.Float(mass), ball.Moment(mass))
	body.SetPosition(vect.Vect{vect.Float(x), windowHeight})
	body.SetAngle(vect.Float(rand.Float32() * 2 * math.Pi))
  body.SetAngularVelocity(float32((rand.Float32()*2 - 1) * 50))

	body.AddShape(ball)
	space.AddBody(body)
	balls = append(balls, ball)
}

// step advances the physics engine and cleans up any balls that are off-screen
func step(dt float32) {
	space.Step(vect.Float(dt))

	for i := 0; i < len(balls); i++ {
		p := balls[i].Body.Position()
		if p.Y < -100 {
			space.RemoveBody(balls[i].Body)
			balls[i] = nil
			balls = append(balls[:i], balls[i+1:]...)
			i-- // consider same index again
		}
	}
}

// createBodies sets up the chipmunk space and static bodies
func createBodies() {
	space = chipmunk.NewSpace()
	space.Gravity = vect.Vect{0, -900}
  space.Iterations = 16
}


func main() {
	runtime.LockOSThread()

	// initialize glfw
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize GLFW: ", err)
	}
	defer glfw.Terminate()

	// create window
  glfw.WindowHint(glfw.Resizable, glfw.True)
  glfw.WindowHint(glfw.ContextVersionMajor, 3)
  glfw.WindowHint(glfw.ContextVersionMinor, 3)
  glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
  glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, os.Args[0], nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	// set up physics
	createBodies()


  program, err := newProgram("vertexShader.vertexshader", "fragmentShader.fragmentshader")
  if err != nil {
    panic(err)
  }
  gl.UseProgram(program)

  model := mgl.Ident4()
  modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
  gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

  vertexBufferData := []float32{}
  sides := int32(60)
  for a := 0.0; a <= 2*math.Pi; a += (2 * math.Pi / float64(sides)) {
    vertexBufferData = append(vertexBufferData, float32(math.Sin(a)))
    vertexBufferData = append(vertexBufferData, float32(math.Cos(a)))
    vertexBufferData = append(vertexBufferData, 0)
  }
  vertexBufferData = append(vertexBufferData, float32(math.Sin(0)))
  vertexBufferData = append(vertexBufferData, float32(math.Cos(0)))
  vertexBufferData = append(vertexBufferData, 0)
  vertexBufferData = append(vertexBufferData, []float32{0.0, 0.0, 0.0}...)

  // Configure the vertex data
  var vao uint32
  gl.GenVertexArrays(1, &vao)
  gl.BindVertexArray(vao)

  var vbo uint32
  gl.GenBuffers(1, &vbo)
  gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
  gl.BufferData(gl.ARRAY_BUFFER, len(vertexBufferData)*4, gl.Ptr(vertexBufferData), gl.STATIC_DRAW)

  vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
  gl.EnableVertexAttribArray(vertAttrib)
  gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

  // Configure global settings
  gl.Enable(gl.DEPTH_TEST)
  gl.DepthFunc(gl.LESS)
  gl.ClearColor(1.0, 1.0, 1.0, 1.0)
  //gl.Viewport(0,0, windowWidth, windowHeight)

  frame := 0
  bigBall := 0
  bump := 0
  count = 0

  font, err = glfont.LoadFont("roboto/Roboto-Light.ttf", int32(52), windowWidth, windowHeight)
  if err != nil {
    log.Panicf("LoadFont: %v", err)
  }
  font.SetColor(0.0, 0.0, 0.0, 1.0) //r,g,b,a font color

	for !window.ShouldClose() {
    frame++
    bigBall++
    bump++
		addBall()
		addBall()
		addBall()
		addBall()
		addBall()
		addBall()
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
    gl.UseProgram(program)
    for _, ball := range balls {
      pos := ball.Body.Position()
      _ = pos
      csA, _ := ball.ShapeClass.(*chipmunk.CircleShape)
      rot := ball.Body.Angle()
      model = mgl.Ortho2D(0, windowWidth, windowHeight, 0).Mul4(mgl.Translate3D(float32(pos.X), windowHeight-float32(pos.Y), 0).Mul4(mgl.HomogRotate3D(float32(rot), mgl.Vec3{0,0,-1}).Mul4(mgl.Scale3D(float32(csA.Radius), float32(csA.Radius), 0))))
      gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
      gl.BindVertexArray(vao)
      gl.DrawArrays(gl.LINE_LOOP, 0, sides + 6)
    }

    if frame == 60 {
      count = len(space.Bodies)
      fmt.Printf("Count: %v\n", count)
      frame = 0
    }
    if bigBall == 60*2 {
      addBigBall()
      bigBall = 0
    }

		step(1.0 / 120.0)

    if bump == 60 {
      staticBody := chipmunk.NewBody(Inf, Inf)
      staticLines = []*chipmunk.Shape{
        chipmunk.NewSegment(vect.Vect{390.0, 100.0}, vect.Vect{1100.0, 400.0}, 20),
        chipmunk.NewSegment(vect.Vect{50, 900.0}, vect.Vect{407.0, 346.0}, 20),
      }
      for _, segment := range staticLines {
        segment.SetElasticity(0.6)
        staticBody.AddShape(segment)
      }
      staticBody.IgnoreGravity = true
      space.AddBody(staticBody)
    } else if bump == 120 && len(space.Bodies) > 4000 {
      space.RemoveBody(staticLines[0].Body)
      staticBody = chipmunk.NewBody(Inf, Inf)
      staticLines = []*chipmunk.Shape{
        chipmunk.NewSegment(vect.Vect{390.0, 100.0}, vect.Vect{1100.0, 400.0}, 20),
      }
      for _, segment := range staticLines {
        segment.SetElasticity(0.6)
        staticBody.AddShape(segment)
      }
      staticBody.IgnoreGravity = true
      space.AddBody(staticBody)
      bump = 0
    } else if bump == 320 {
      bump = 60
    }
    font.Printf(10, 100, 0.3, "Count: %v", count) //x,y,scale,string,printf args

		window.SwapBuffers()
		glfw.PollEvents()

	}
}

func newProgram(vertexFilePath, fragmentFilePath string) (uint32, error) {

  // Load both shaders
  vertexShaderID, fragmentShaderID, err := loadShaders(vertexFilePath, fragmentFilePath)
  if err != nil {
    return 0, err
  }

  // Create new program
  programID := gl.CreateProgram()
  gl.AttachShader(programID, vertexShaderID)
  gl.AttachShader(programID, fragmentShaderID)
  gl.LinkProgram(programID)

  // Check status of program
  var status int32
  gl.GetProgramiv(programID, gl.LINK_STATUS, &status)
  if status == gl.FALSE {
    var logLength int32
    gl.GetProgramiv(programID, gl.INFO_LOG_LENGTH, &logLength)

    log := strings.Repeat("\x00", int(logLength+1))
    gl.GetProgramInfoLog(programID, logLength, nil, gl.Str(log))

    return 0, fmt.Errorf("failed to link program: %v", log)
  }

  // Detach shaders
  gl.DetachShader(programID, vertexShaderID)
  gl.DetachShader(programID, fragmentShaderID)

  // Delete shaders
  gl.DeleteShader(vertexShaderID)
  gl.DeleteShader(fragmentShaderID)

  return programID, nil
}

func loadShaders(vertexFilePath, fragmentFilePath string) (uint32, uint32, error) {

  // Compile vertex shader
  vertexShaderID, err := compileShader(readShaderCode(vertexFilePath), gl.VERTEX_SHADER)
  if err != nil {
    return 0, 0, err
  }

  // Compile fragment shader
  fragmentShaderID, err := compileShader(readShaderCode(fragmentFilePath), gl.FRAGMENT_SHADER)
  if err != nil {
    return 0, 0, err
  }

  return vertexShaderID, fragmentShaderID, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {

  // Create new shader 
  shader := gl.CreateShader(shaderType)
  // Convert shader string to null terminated c string
  shaderCode, free := gl.Strs(source)
  defer free()
  gl.ShaderSource(shader, 1, shaderCode, nil)

  // Compile shader
  gl.CompileShader(shader)

  // Check shader status
  var status int32
  gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
  if status == gl.FALSE {
    var logLength int32
    gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

    log := strings.Repeat("\x00", int(logLength+1))
    gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

    return 0, fmt.Errorf("failed to compile %v: %v", source, log)
  }
  return shader, nil
}

func readShaderCode(filePath string) string {
  code := ""
  f, err := os.Open(filePath)
  if err != nil {
    log.Fatal(err)
  }
  defer f.Close()

  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    code += "\n" + scanner.Text()
  }
  if err := scanner.Err(); err != nil {
    log.Fatal(err)
  }
  code += "\x00"
  return code
}
