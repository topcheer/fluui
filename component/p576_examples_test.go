package component_test

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/topcheer/fluui/component"
)

// p576_examples.go — Second batch of missing Example functions (Direction E).

func Example_aiMetricsCard() {
	c := component.NewAIMetricsCard()
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 4})
	fmt.Println("ok")
	// Output: ok
}

func Example_approvalConfirm() {
	d := component.NewApprovalConfirmDialog("Deploy", "Deploy to production?")
	fmt.Printf("Title:%s\n", d.Title())
	// Output: Title:Deploy
}

func Example_criticalBadge() {
	component.NewCriticalBadge("CRITICAL")
	fmt.Println("ok")
	// Output: ok
}

func Example_errorBadge() {
	component.NewErrorBadge("FAILED")
	fmt.Println("ok")
	// Output: ok
}

func Example_infoBadge() {
	component.NewInfoBadge("INFO")
	fmt.Println("ok")
	// Output: ok
}

func Example_neutralBadge() {
	component.NewNeutralBadge("NOTE")
	fmt.Println("ok")
	// Output: ok
}

func Example_confirmDialog() {
	d := component.NewConfirmDialog("Delete", "Are you sure?")
	fmt.Printf("Title:%s\n", d.Title())
	// Output: Title:Delete
}

func Example_infoDialog() {
	d := component.NewInfoDialog("About", "Fluui v1.0")
	fmt.Printf("Title:%s\n", d.Title())
	// Output: Title:About
}

func Example_promptDialog() {
	d := component.NewPromptDialog("Enter name", "Name:", "default")
	fmt.Printf("Title:%s\n", d.Title())
	// Output: Title:Enter name
}

func Example_questionnaireDialog() {
	d := component.NewQuestionnaireDialog("Survey", nil)
	fmt.Printf("Title:%s\n", d.Title())
	// Output: Title:Survey
}

func Example_filePicker() {
	fp := component.NewFilePicker("/tmp")
	fp.SetFilter("go")
	fmt.Println("ok")
	// Output: ok
}

func Example_markdownViewer() {
	component.NewMarkdownViewer("# Hello\nWorld")
	fmt.Println("ok")
	// Output: ok
}

func Example_promptTemplateTree() {
	t := component.NewPromptTemplateTree()
	t.AddNode(0, "root", "Hello {{name}}!", true)
	fmt.Printf("Nodes:%d\n", t.NodeCount())
	// Output: Nodes:1
}

func ExampleRule() {
	r := component.NewRule()
	r.SetChar('=')
	fmt.Println("ok")
	// Output: ok
}

func Example_separator() {
	component.NewSeparator()
	fmt.Println("ok")
	// Output: ok
}

func Example_searchFilter() {
	sf := component.NewSearchFilter()
	sf.SetQuery("test")
	fmt.Printf("Query:%s\n", sf.Query())
	// Output: Query:test
}

func ExampleReactive() {
	r := component.NewReactive("initial")
	fmt.Printf("Value:%v\n", r.Get())
	// Output: Value:initial
}

func Example_reactiveWithDirty() {
	var dirty atomic.Bool
	r := component.NewReactiveWithDirty("val", &dirty)
	r.Set("newval")
	fmt.Printf("Dirty:%v\n", dirty.Load())
	// Output: Dirty:true
}

func ExamplePretty() {
	component.NewPretty(42)
	fmt.Println("ok")
	// Output: ok
}

func Example_prettyString() {
	component.NewPrettyString("hello")
	fmt.Println("ok")
	// Output: ok
}

func Example_calendarWithDate() {
	component.NewCalendarWithDate(time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))
	fmt.Println("ok")
	// Output: ok
}

func Example_buttonWithVariant() {
	b := component.NewButtonWithVariant("Click", component.ButtonPrimary)
	fmt.Printf("Label:%s\n", b.Label())
	// Output: Label:Click
}

func Example_badgeWithSize() {
	component.NewBadgeWithSize("NEW", component.BadgeInfo, component.BadgeSizeNormal)
	fmt.Println("ok")
	// Output: ok
}

func Example_checkboxField() {
	c := component.NewCheckboxField("Accept", "accept", true)
	fmt.Printf("Label:%s\n", c.Label())
	// Output: Label:Accept
}

func Example_selectField() {
	s := component.NewSelectField("Color", "color", []string{"Red", "Green", "Blue"})
	fmt.Printf("Label:%s\n", s.Label())
	// Output: Label:Color
}

func Example_reactiveField() {
	var dirty atomic.Bool
	rf := component.NewReactiveField("default", &dirty)
	fmt.Printf("Value:%v\n", rf.Get())
	// Output: Value:default
}

func Example_dialogButton() {
	component.NewDialogButton("OK", component.DialogResultOK)
	fmt.Println("ok")
	// Output: ok
}

func Example_menuItem() {
	mi := component.NewMenuItem("file", "File")
	fmt.Printf("Label:%s\n", mi.Label)
	// Output: Label:File
}

func Example_canvasLayer() {
	component.NewCanvasLayer("bg", 80, 24)
	fmt.Println("ok")
	// Output: ok
}

func ExampleClear() {
	component.NewClear()
	fmt.Println("ok")
	// Output: ok
}
