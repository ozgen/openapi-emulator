package samples

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ozgen/openapi-sample-emulator/logger"
	"github.com/sirupsen/logrus"
)

type ScenarioResolver interface {
	ResolveScenarioFile(
		scenarioPath string,
		sc *Scenario,
		method string,
		swaggerTpl string,
		actualPath string,
	) (file string, state string, err error)
}

type Scenario struct {
	Version int    `json:"version"`
	Mode    string `json:"mode"` // "step" | "time"
	Key     struct {
		PathParam string `json:"pathParam"`
	} `json:"key"`

	Sequence []ScenarioEntry `json:"sequence,omitempty"` // step mode
	Timeline []TimelineEntry `json:"timeline,omitempty"` // time mode

	Behavior Behavior `json:"behavior"`
}

type ScenarioEntry struct {
	State string `json:"state"`
	File  string `json:"file"`
}

type TimelineEntry struct {
	AfterMs int64  `json:"afterMs"`
	State   string `json:"state"`
	File    string `json:"file"`
}

type Behavior struct {
	AdvanceOn  []MatchRule `json:"advanceOn,omitempty"`
	ResetOn    []MatchRule `json:"resetOn,omitempty"`
	StartOn    []MatchRule `json:"startOn,omitempty"`
	RepeatLast bool        `json:"repeatLast"`
}

type MatchRule struct {
	Method string `json:"method"`
	Path   string `json:"path,omitempty"`
}

// ScenarioEngine holds runtime state (in-memory, v1).
type ScenarioEngine struct {
	mu sync.Mutex

	// per (scenarioFilePath + "::" + keyVal)
	stepIndex map[string]int
	startedAt map[string]time.Time
	log       *logrus.Logger
}

func NewScenarioEngine() *ScenarioEngine {
	return &ScenarioEngine{
		stepIndex: map[string]int{},
		startedAt: map[string]time.Time{},
		log:       logger.GetLogger(),
	}
}

func LoadScenario(scenarioPath string) (*Scenario, error) {
	log := logger.GetLogger()
	b, err := os.ReadFile(scenarioPath)
	if err != nil {
		return nil, err
	}
	var sc Scenario
	if err := json.Unmarshal(b, &sc); err != nil {
		log.WithError(err).Error("failed to load scenario")
		return nil, fmt.Errorf("parse scenario.json: %w", err)
	}

	if sc.Version != 1 {
		log.WithField("version", sc.Version).Error("invalid scenario version," +
			" version should be 1")
		return nil, fmt.Errorf("unsupported scenario version: %d", sc.Version)
	}
	sc.Mode = strings.TrimSpace(sc.Mode)
	if sc.Mode != "step" && sc.Mode != "time" {
		log.WithField("mode", sc.Mode).Errorf("invalid mode :%s", sc.Mode)
		return nil, fmt.Errorf("invalid scenario mode: %q", sc.Mode)
	}
	if strings.TrimSpace(sc.Key.PathParam) == "" {
		log.WithField("pathParam", sc.Key.PathParam).Error("pathParam is required")
		return nil, fmt.Errorf("scenario.key.pathParam is required")
	}

	return &sc, nil
}

func (e *ScenarioEngine) ResolveScenarioFile(
	scenarioPath string,
	sc *Scenario,
	method string,
	swaggerTpl string,
	actualPath string,
) (file string, state string, err error) {
	method = strings.ToUpper(method)

	keyVal, ok := extractPathParam(swaggerTpl, actualPath, sc.Key.PathParam)
	if !ok || strings.TrimSpace(keyVal) == "" {
		e.log.WithError(err).Errorf("failed to resolve path param %q", actualPath)
		return "", "", fmt.Errorf("cannot extract key path param %q from path %q using template %q",
			sc.Key.PathParam, actualPath, swaggerTpl)
	}

	// apply reset rules (best-effort)
	if matchesAny(sc.Behavior.ResetOn, method, actualPath) {
		e.mu.Lock()
		delete(e.stepIndex, scenarioKey(scenarioPath, keyVal))
		delete(e.startedAt, scenarioKey(scenarioPath, keyVal))
		e.mu.Unlock()
	}

	switch sc.Mode {
	case "step":
		return e.resolveStep(scenarioPath, sc, method, keyVal)
	case "time":
		return e.resolveTime(scenarioPath, sc, method, actualPath, keyVal)
	default:
		e.log.WithField("mode", sc.Mode).Errorf("invalid mode :%s", sc.Mode)
		return "", "", fmt.Errorf("unsupported mode %q", sc.Mode)
	}
}

func (e *ScenarioEngine) resolveStep(scenarioPath string, sc *Scenario, method string, keyVal string) (string, string, error) {
	if len(sc.Sequence) == 0 {
		e.log.WithField("path", scenarioPath).Info("no step sequence")
		return "", "", fmt.Errorf("step mode requires non-empty sequence")
	}

	k := scenarioKey(scenarioPath, keyVal)

	e.mu.Lock()
	idx := e.stepIndex[k]
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sc.Sequence) {
		idx = len(sc.Sequence) - 1
	}
	entry := sc.Sequence[idx]

	// advance AFTER selecting current entry (so 1st GET returns first state)
	if matchesAny(sc.Behavior.AdvanceOn, method, "") {
		next := idx + 1
		if next >= len(sc.Sequence) {
			if sc.Behavior.RepeatLast {
				next = len(sc.Sequence) - 1
			} else {
				next = len(sc.Sequence) - 1
			}
		}
		e.stepIndex[k] = next
	} else {
		e.stepIndex[k] = idx
	}
	e.mu.Unlock()

	return entry.File, entry.State, nil
}

func (e *ScenarioEngine) resolveTime(scenarioPath string, sc *Scenario, method string, actualPath string, keyVal string) (string, string, error) {
	if len(sc.Timeline) == 0 {
		e.log.WithField("path", scenarioPath).Info("no timeline")
		return "", "", fmt.Errorf("time mode requires non-empty timeline")
	}

	k := scenarioKey(scenarioPath, keyVal)

	e.mu.Lock()
	t0, ok := e.startedAt[k]
	if !ok {
		if len(sc.Behavior.StartOn) == 0 || matchesAny(sc.Behavior.StartOn, method, actualPath) {
			t0 = time.Now()
			e.startedAt[k] = t0
		} else {
			t0 = time.Now()
			e.startedAt[k] = t0
		}
	}
	elapsed := time.Since(t0)
	elapsedMs := elapsed.Milliseconds()
	e.mu.Unlock()

	chosen := sc.Timeline[0]
	for _, t := range sc.Timeline {
		if t.AfterMs <= elapsedMs {
			chosen = t
		} else {
			break
		}
	}

	return chosen.File, chosen.State, nil
}

func scenarioKey(scenarioPath, keyVal string) string {
	return scenarioPath + "::" + keyVal
}

func matchesAny(rules []MatchRule, method string, actualPath string) bool {
	method = strings.ToUpper(method)
	for _, r := range rules {
		if strings.ToUpper(strings.TrimSpace(r.Method)) != method {
			continue
		}
		p := strings.TrimSpace(r.Path)
		if p == "" {
			return true
		}
		if matchTemplatePath(p, actualPath) {
			return true
		}
	}
	return false
}

func matchTemplatePath(tpl, actual string) bool {
	tplParts := strings.Split(strings.Trim(tpl, "/"), "/")
	actParts := strings.Split(strings.Trim(actual, "/"), "/")
	if len(tplParts) != len(actParts) {
		return false
	}
	for i := range tplParts {
		t := tplParts[i]
		a := actParts[i]
		if strings.HasPrefix(t, "{") && strings.HasSuffix(t, "}") {
			continue
		}
		if t != a {
			return false
		}
	}
	return true
}

func extractPathParam(swaggerTpl, actualPath, want string) (string, bool) {
	tplParts := strings.Split(strings.Trim(swaggerTpl, "/"), "/")
	actParts := strings.Split(strings.Trim(actualPath, "/"), "/")
	if len(tplParts) != len(actParts) {
		return "", false
	}
	for i := range tplParts {
		p := tplParts[i]
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(p, "{"), "}")
			if name == want {
				return actParts[i], true
			}
		}
	}
	return "", false
}

func ScenarioPathForSwagger(baseDir, swaggerPath, filename string) string {
	pathDir := strings.TrimPrefix(swaggerPath, "/")
	pathDir = filepath.FromSlash(pathDir)
	return filepath.Join(baseDir, pathDir, filename)
}
