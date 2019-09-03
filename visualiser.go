/* See also examples on https://github.com/golang-ui/nuklear */
package main

import (
	"fmt"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/sindbach/nuklear/nk"
	"github.com/xlab/closer"
	"reflect"
	"runtime"
	"time"
)

const (
	winWidth         = 1024
	winHeight        = 800
	maxVertexBuffer  = 100 * 1024 * 1024
	maxElementBuffer = 100 * 1024 * 1024
)

var (
	cTEXT           = nk.NkRgb(185, 185, 185)
	cLINE           = nk.NkRgb(95, 95, 95)
	cTXLINE         = nk.NkRgb(255, 255, 255)
	cGENISIS        = nk.NkRgb(0, 204, 255)
	cFINALISED      = nk.NkRgb(204, 255, 0)
	cFINALISEDOTHER = nk.NkRgb(180, 180, 180)
	cINVALID        = nk.NkRgb(255, 51, 0)
	cCANONICAL      = nk.NkRgb(243, 243, 21)
	cSTALE          = nk.NkRgb(194, 14, 213)
	cPRUNED         = nk.NkRgb(95, 95, 95)
)

func init() {
	runtime.LockOSThread()
}

type Visualiser struct {
	shards            []Shard
	channels          Communication
	state             Command
	viewShard         int32
	font              *nk.UserFont
	offSetX           float64
	scaleX            float64
	selectedNode      *ChainBlock
	selectedNodeChain int
	selectedTX        *Transaction
}

type Coordinate struct {
	x     float32
	y     float32
	color nk.Color
}

func (visualiser *Visualiser) Init() {
}

func (visualiser *Visualiser) Run() {
	////////////////////////////////////////////
	if err := glfw.Init(); err != nil {
		closer.Fatalln(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(winWidth, winHeight, "Smart-Shard Visualiser", nil, nil)
	if err != nil {
		closer.Fatalln(err)
	}
	win.MakeContextCurrent()

	width, height := win.GetSize()

	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: initialisation failed:", err)
	}
	gl.Viewport(0, 0, int32(width), int32(height))
	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	win.SetScrollCallback(func(win *glfw.Window, x, y float64) {
		visualiser.offSetX = visualiser.offSetX + x
		visualiser.scaleX = visualiser.scaleX - (y / 100)
		if visualiser.scaleX < 0.5 {
			visualiser.scaleX = 0.5
		}
	})

	// Fonts
	fontHandle := initFont(ctx)
	visualiser.font = fontHandle

	// Style
	//ctx.Style().Window().

	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	refreshC := make(chan bool, 1)

	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	refreshC <- true
	for {
		select {
		case <-exitC:
			nk.NkPlatformShutdown()
			glfw.Terminate()
			visualiser.channels.broadCastCommand(Exit)
			close(doneC)
			close(refreshC)
			return
		case <-refreshC:
			if win.ShouldClose() {
				close(exitC)
				continue
			}
			glfw.PollEvents()
			visualiser.gfxMain(win, ctx)
			//time.Sleep(time.Second/10)
			refreshC <- true
		}
	}
	runtime.KeepAlive(fontHandle)
}

func (visualiser *Visualiser) gfxMain(win *glfw.Window, ctx *nk.Context) {

	// Create GUI frame
	nk.NkPlatformNewFrame()
	width, height := win.GetSize()

	// Update visualisation
	shard := visualiser.viewShard + 1
	visualiser.shards[shard].UpdateVisualisation()

	bounds := nk.NkRect(0, 0, float32(width), float32(height))
	update := nk.NkBegin(ctx, "Shard Inspector", bounds, 0)

	if update > 0 {
		visualiser.drawLayout(win, ctx)
	}
	nk.NkEnd(ctx)

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, nk.NkRgba(50, 50, 50, 255))
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func initFont(ctx *nk.Context) *nk.UserFont {
	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		sansFontHandle := sansFont.Handle()
		nk.NkStyleSetFont(ctx, sansFontHandle)
		return sansFontHandle
	}
	return nil
}

func (visualiser *Visualiser) drawLayout(win *glfw.Window, ctx *nk.Context) {

	_, height := win.GetSize()
	canvas := nk.NkWindowGetCanvas(ctx)

	toolbarHeight := float32(25)
	paddingX := float32(15)
	paddingY := float32(10)

	// Draw Menu
	nk.NkLayoutRowStatic(ctx, toolbarHeight, 120, 8)
	{
		// Start button
		if nk.NkButtonLabel(ctx, "Start") > 0 {
			visualiser.start()
		}

		if nk.NkButtonLabel(ctx, "Pause") > 0 {
			visualiser.stop()
		}

		if nk.NkButtonLabel(ctx, "Pretty Print") > 0 {
			visualiser.prettyPrint()
		}

		comboString := ""
		for i := 1; i <= ShardCount; i++ {
			comboString = fmt.Sprint(comboString, fmt.Sprintf("Shard %d", i), "\x00")
		}
		nk.NkComboboxString(ctx, comboString, &visualiser.viewShard, ShardCount, 25, nk.NkVec2(150, 200))

		forkProb := 1 - ProbabilityBuildOnLongestChain
		nk.NkLabel(ctx, fmt.Sprintf("Forks (%.0f%%):", forkProb*100), nk.TextAlignRight|nk.TextAlignMiddle)
		newForkProb := nk.NkSlideFloat(ctx, 0, float32(forkProb), 0.5, 0.1)
		if newForkProb != float32(forkProb) {
			ProbabilityBuildOnLongestChain = float64(1 - newForkProb)
		}

		nk.NkLabel(ctx,"Finalise Speed:", nk.TextAlignRight|nk.TextAlignMiddle)
		newSpeed := nk.NkSlideFloat(ctx, 1, float32(FinalisationPeriod.min), 8, 1)
		if newSpeed != float32(FinalisationPeriod.min) {
			FinalisationPeriod.min = int(newSpeed)
			FinalisationPeriod.max = int(newSpeed) + 2
		}

	}

	winStartX := paddingX
	winStartY := toolbarHeight + 2*paddingY

	// Calculate  width
	t := time.Now()
	duration := t.Sub(StartTime)
	winWidth := duration.Seconds()*pixelsPerSecond + 2*float64(paddingX)

	shard := visualiser.viewShard + 1

	nk.NkLayoutRowTemplateBegin(ctx, float32(height)-30-toolbarHeight)
	nk.NkLayoutRowTemplatePushVariable(ctx, 320)
	nk.NkLayoutRowTemplatePushStatic(ctx, 210)
	nk.NkLayoutRowTemplateEnd(ctx)

	if nk.NkGroupBegin(ctx, "", 0) > 0 {

		nk.NkLayoutRowDynamic(ctx, float32(150), 1)

		winStartX = 0 + float32(visualiser.offSetX) + 80
		winStartY = 115

		for i := 1; i <= ShardCount; i++ {

			widthX := float32(winWidth)
			widthY := float32(80)

			// Plot chain
			genesisBlock := visualiser.shards[shard].chains[i].genesisBlock
			visualiser.drawChain(ctx, canvas, i, winStartX, winStartY, widthX, widthY, genesisBlock)

			winStartY += 155
		}

		nk.NkLayoutRowDynamic(ctx, 25, 1)
		text := string("Tip: Use mouse scroll to scale and move the x-as of block trees.")
		nk.NkText(ctx, text, int32(len(text)), nk.TextAlignLeft)

		if visualiser.selectedNode != nil {

			for _, tx := range visualiser.selectedNode.block.TXIn {
				blocksOut := visualiser.shards[shard].chains[tx.SourceShard].GetChainBlock(tx)

				for _, block := range blocksOut {

					x0 := winStartX + (block.coordinate.x * float32(visualiser.scaleX)) + 5
					y0 := 115 + float32(155*tx.SourceShard) - 155 + block.coordinate.y + 5
					x1 := winStartX + (visualiser.selectedNode.coordinate.x * float32(visualiser.scaleX)) + 5
					y1 := 115 + float32(155*tx.TargetShard) - 155 + visualiser.selectedNode.coordinate.y + 5

					nk.NkStrokeLine(canvas, x0, y0, x1, y1, 1.0, cTXLINE)

				}
			}
		}
		nk.NkGroupEnd(ctx)
	}
	if nk.NkGroupBegin(ctx, "", 0) > 0 {
		visualiser.drawBockInspector(ctx)
		visualiser.drawTXInspector(ctx)
		nk.NkGroupEnd(ctx)
	}
}

func (visualiser *Visualiser) drawBockInspector(ctx *nk.Context) {

	nk.NkLayoutRowDynamic(ctx, float32(400), 1)

	if visualiser.selectedNode != nil {

		nk.NkGroupBegin(ctx, "Block Inspector", nk.WindowTitle|nk.WindowBorder)

		nk.NkLayoutRowDynamic(ctx, 25, 1)

		nk.NkLabelColored(ctx, "Hash:", nk.TextAlignCentered|nk.TextAlignMiddle, cTXLINE)
		nk.NkLabel(ctx, fmt.Sprintf(" %x", visualiser.selectedNode.block.Hash), nk.TextAlignLeft|nk.TextAlignMiddle)

		nk.NkLabelColored(ctx, "TX - IN:", nk.TextAlignCentered|nk.TextAlignMiddle, cTXLINE)

		for _, tx := range visualiser.selectedNode.block.TXIn {
			if nk.NkSelectLabel(ctx, fmt.Sprintf(" %x", tx.Hash), nk.TextAlignLeft|nk.TextAlignMiddle, visualiser.isSelectedTx(tx)) > 0 {
				visualiser.selectedTX = tx
			}
		}

		nk.NkLabelColored(ctx, "TX - OUT:", nk.TextAlignCentered|nk.TextAlignMiddle, cTXLINE)

		for _, tx := range visualiser.selectedNode.block.TXOut {
			if nk.NkSelectLabel(ctx, fmt.Sprintf("%x", tx.Hash), nk.TextAlignLeft|nk.TextAlignMiddle, visualiser.isSelectedTx(tx)) > 0 {
				visualiser.selectedTX = tx
			}
		}

		nk.NkGroupEnd(ctx)
	}
}

func (visualiser *Visualiser) isSelectedTx(transaction *Transaction) int32 {
	if visualiser.selectedTX != nil && visualiser.selectedTX.Hash == transaction.Hash {
		return 1
	}
	return 0
}

func (visualiser *Visualiser) drawTXInspector(ctx *nk.Context) {

	nk.NkLayoutRowDynamic(ctx, float32(230), 1)

	if visualiser.selectedTX != nil {

		nk.NkGroupBegin(ctx, "TX Inspector", nk.WindowTitle|nk.WindowBorder)

		nk.NkLayoutRowDynamic(ctx, 25, 1)

		nk.NkLabelColored(ctx, "Hash:", nk.TextAlignCentered|nk.TextAlignMiddle, cTXLINE)
		nk.NkLabel(ctx, fmt.Sprintf(" %x", visualiser.selectedTX.Hash), nk.TextAlignLeft|nk.TextAlignMiddle)

		nk.NkLabelColored(ctx, "Source Shard:", nk.TextAlignCentered|nk.TextAlignMiddle, cTXLINE)
		nk.NkLabel(ctx, fmt.Sprintf(" %x", visualiser.selectedTX.SourceShard), nk.TextAlignCentered|nk.TextAlignMiddle)

		nk.NkLabelColored(ctx, "Target Shard:", nk.TextAlignCentered|nk.TextAlignMiddle, cTXLINE)
		nk.NkLabel(ctx, fmt.Sprintf(" %x", visualiser.selectedTX.TargetShard), nk.TextAlignCentered|nk.TextAlignMiddle)

		nk.NkGroupEnd(ctx)
	}
}

func (visualiser *Visualiser) drawChain(ctx *nk.Context, canvas *nk.CommandBuffer, shard int, winStartX float32, winStartY float32, width float32, height float32, genisisBlock *ChainBlock) {

	nk.NkGroupBegin(ctx, fmt.Sprintf("Shard %d", shard), nk.WindowBorder|nk.WindowTitle)
	input := ctx.Input()
	visualiser.drawBlock(canvas, input, shard, winStartX, winStartY, genisisBlock)
	nk.NkGroupEnd(ctx)
}

func (visualiser *Visualiser) drawBlock(canvas *nk.CommandBuffer, input *nk.Input, shard int, winStartX float32, winStartY float32, chainBlock *ChainBlock) {

	// Draw lines child blocks
	for _, child := range chainBlock.children {
		x0 := winStartX + (chainBlock.coordinate.x * float32(visualiser.scaleX)) + 5
		y0 := winStartY + chainBlock.coordinate.y + 5
		x1 := winStartX + (child.coordinate.x * float32(visualiser.scaleX)) + 5
		y1 := winStartY + child.coordinate.y + 5

		nk.NkStrokeLine(canvas, x0, y0, x1, y1, 1.0, cLINE)
	}

	// Draw current block
	x := winStartX + (chainBlock.coordinate.x * float32(visualiser.scaleX))
	y := winStartY + chainBlock.coordinate.y

	c1 := nk.NkRect(x, y, 10.0, 10.0)
	if visualiser.selectedNode != nil && reflect.DeepEqual(chainBlock.block.Hash, visualiser.selectedNode.block.Hash) {
		nk.NkFillCircle(canvas, c1, chainBlock.coordinate.color)
		nk.NkStrokeCircle(canvas, c1, 2, chainBlock.coordinate.color)
	} else {
		nk.NkStrokeCircle(canvas, c1, 2, chainBlock.coordinate.color)
	}

	if nk.NkInputHasMouseClickDownInRect(input, nk.ButtonLeft, c1, 1) > 0 {
		visualiser.selectedNode = chainBlock
		visualiser.selectedNodeChain = shard
	}

	// Draw child blocks
	for _, child := range chainBlock.children {
		visualiser.drawBlock(canvas, input, shard, winStartX, winStartY, child)
	}
}

func (visualiser *Visualiser) start() {
	visualiser.state = Run
	visualiser.channels.broadCastCommand(Run)
	fmt.Println("START")
}

func (visualiser *Visualiser) stop() {
	visualiser.state = Pause
	visualiser.channels.broadCastCommand(Pause)
	fmt.Println("PAUSE")
}

func (visualiser *Visualiser) prettyPrint() {
	shard := visualiser.viewShard + 1
	visualiser.shards[shard].chains[shard].PrettyPrint()
}
