// Package main is a test program for the grafana explore page
package main

import (
	"log/slog"
	"time"

	"github.com/go-rod/rod"
)

type grafana struct {
	BaseURL string
	*rod.Browser
	*rod.Page
}

type grafanaExplore struct {
	grafana
}

func (g *grafana) openExplore() *grafanaExplore {
	slog.Info("openExplore")
	g.Browser = rod.New().MustConnect()
	g.Page = g.Browser.MustPage(g.BaseURL + "/explore").MustWindowFullscreen().Timeout(5 * time.Second).MustWaitStable().CancelTimeout()
	return &grafanaExplore{*g}
}

func (g *grafanaExplore) selectDatasource(datasourceName string) *grafanaExplore {
	slog.Info("selectDatasource: input", "datasourceName", datasourceName)
	g.Page.MustElement("#data-source-picker").MustInput(datasourceName).Timeout(5 * time.Second).MustWaitStable().CancelTimeout()

	// there should only be one matching datasource
	slog.Info("selectDatasource: click")
	button := g.Page.MustElement(`div[data-testid="data-source-card"]`).MustElement("button")
	button.MustClick()
	return g
}

type lokiEditor int

const (
	lokiQueryEditor lokiEditor = iota
	lokiCodeEditor
)

func (g *grafanaExplore) selectLokiEditor(editorMode lokiEditor) {
	slog.Info("selectLokiEditor", "editorMode", editorMode)
	modes := g.Page.MustElement(`div[data-testid="QueryEditorModeToggle"]`).MustElements(`div[data-testid="data-testid radio-button"]`)
	switch editorMode {
	case lokiQueryEditor:
		modes.First().MustElement("input").MustClick()
	case lokiCodeEditor:
		modes.Last().MustElement("input").MustClick()
	}
}

func (g *grafanaExplore) inputRunQuery(query string) {
	slog.Info("inputQuery", "query", query)
	g.Page.MustElement(`div[aria-label="Query field"]`).MustElement(`div.view-lines.monaco-mouse-cursor-text`).MustClick().MustInput(query)

	slog.Info("run query")
	g.Page.MustElement(`button[aria-label="Run query"]`).MustClick()
}

func main() {
	g := grafana{BaseURL: "http://localhost:3000"}
	explore := g.openExplore()
	explore.MustSetViewport(2048, 1200, 1, false)
	explore.selectDatasource("Loki").selectLokiEditor(lokiCodeEditor)
	explore.inputRunQuery(`{app="boomer"} | logfmt |= "boom-child"`)

	slog.Info("wait for query to finish")
	explore.MustWaitStable()

	// find the log scroll view, and scroll it up a bit
	logScroll := explore.MustElement(`div[data-testid="data-testid explorer scroll view"]`).MustElement(`div.scrollbar-view`)
	logScroll.MustEval("() => this.scrollBy(0, 600)")

	// expand the first log
	row := logScroll.MustElement("table").MustElement("tr")
	row.MustClick()
	explore.MustWaitStable()

	// select the log details
	details, _ := row.Next()

	// click the "View Trace" button
	details.MustElementR("span", "View Trace").MustClick()
	logScroll.MustEval("() => this.scrollBy(0, 600)")
	explore.MustWaitStable().MustScreenshot("logToTrace.png")

	details.MustScreenshot("logDetails.png")

	// click the "Logs to metrics" button
	details.MustElementR("span", "Logs to metrics").MustClick()
	logScroll.MustEval("() => this.scrollBy(0, 600)")
	explore.MustWaitStable().MustScreenshot("logsToMetrics.png")
}
