package ui

import (
	"os"

	"github.com/gertjaap/vertcoin/logging"
	"github.com/therecipe/qt/core"

	"github.com/therecipe/qt/uitools"

	"github.com/therecipe/qt/widgets"
)

func RunUI(closeChannel chan bool) {
	core.QCoreApplication_SetAttribute(core.Qt__AA_ShareOpenGLContexts, true)

	widgets.NewQApplication(len(os.Args), os.Args)
	window := widgets.NewQMainWindow(nil, core.Qt__FramelessWindowHint)

	loader := uitools.NewQUiLoader(window)
	file := core.NewQFile2(":/qml/mainwindow.ui")
	if ok := file.Open(core.QIODevice__ReadOnly); !ok {
		logging.Error("Couldn't load UI")
		return
	}
	defer file.Close()
	widget := loader.Load(file, window)

	widget.SetWindowTitle("Vertcoin")
	widget.Show()
	widgets.QApplication_Exec()
	closeChannel <- true
}
