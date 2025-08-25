package exitcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/passes/inspect"
)

// TestAnalyzer проверяет, что анализатор может быть создан и имеет правильные свойства.
func TestAnalyzer(t *testing.T) {
	// Проверяем, что анализатор имеет правильное имя
	if Analyzer.Name != "exitcheck" {
		t.Errorf("Expected analyzer name to be 'exitcheck', got '%s'", Analyzer.Name)
	}

	// Проверяем, что анализатор требует анализатор inspect
	if len(Analyzer.Requires) != 1 {
		t.Errorf("Expected analyzer to require 1 analyzer, got %d", len(Analyzer.Requires))
	}

	if Analyzer.Requires[0] != inspect.Analyzer {
		t.Errorf("Expected analyzer to require inspect.Analyzer")
	}

	// Проверяем, что анализатор имеет функцию run
	if Analyzer.Run == nil {
		t.Errorf("Expected analyzer to have a run function")
	}
}

// TestAnalyzerDoc проверяет, что анализатор имеет правильную документацию.
func TestAnalyzerDoc(t *testing.T) {
	if Analyzer.Doc == "" {
		t.Errorf("Expected analyzer to have documentation")
	}
}
