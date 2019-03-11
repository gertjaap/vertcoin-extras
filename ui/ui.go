package ui

import (
	"fmt"
	"os"

	"github.com/therecipe/qt/core"

	"github.com/therecipe/qt/uitools"

	"github.com/therecipe/qt/widgets"
)

func RunUI(closeChannel chan bool) {
	core.QCoreApplication_SetAttribute(core.Qt__AA_ShareOpenGLContexts, true)

	widgets.NewQApplication(len(os.Args), os.Args)
	window := widgets.NewQMainWindow(nil, core.Qt__FramelessWindowHint)

	loader := uitools.NewQUiLoader(window)
	file := core.NewQFile2("./ui/mainwindow.ui")
	if ok := file.Open(core.QIODevice__ReadOnly); !ok {
		fmt.Printf("Couldn't load UI")
		return
	}
	defer file.Close()
	widget := loader.Load(file, window)

	var (
		uiButton  = widgets.NewQPushButtonFromPointer(widget.FindChild("pushButton", core.Qt__FindChildrenRecursively).Pointer())
		uiTextBox = widgets.NewQTextEditFromPointer(widget.FindChild("textEdit", core.Qt__FindChildrenRecursively).Pointer())
	)

	if uiButton == nil {
		fmt.Printf("Couldn't find QPushButton!")
	}

	if uiTextBox == nil {
		fmt.Printf("Couldn't find QTextEdit!")
	}

	uiButton.ConnectPressed(func() {
		fmt.Printf("Value: %s", uiTextBox.ToPlainText())
	})
	widget.SetWindowTitle("Vertcoin")
	widget.Show()
	widgets.QApplication_Exec()
	closeChannel <- true
}
