import pytest
from knowledge.generation.validator import Validator

def test_validate_basic():
    content = "---\ntitle: test\n---\nValid content."
    results = Validator.validate(content)
    assert len(results["errors"]) == 0
    assert len(results["warnings"]) == 0

def test_validate_missing_frontmatter():
    content = "Missing frontmatter."
    results = Validator.validate(content)
    assert "Missing frontmatter block" in results["errors"]

def test_validate_banned_phrases():
    content = "---\ntitle: test\n---\nHere is the document for you."
    results = Validator.validate(content)
    assert any("Conversational phrase detected" in err for err in results["warnings"])

def test_validate_banned_words():
    content = "---\ntitle: test\n---\nWe will utilize this feature."
    results = Validator.validate(content)
    assert any("Banned word detected" in err for err in results["warnings"])

def test_validate_nobleforge():
    content = "---\ntitle: test\n---\nContact NobleForge for help."
    results = Validator.validate(content)
    assert any('Banned term detected: "NobleForge"' in err for err in results["warnings"])


def test_validate_code_block_delegation_success(monkeypatch):
    content = "---\ntitle: test\n---\n```python\nprint('hello')\n```"
    
    # Mock the PythonAdapter
    class MockAdapter:
        def validate_content(self, code):
            return []
            
    monkeypatch.setattr(Validator, 'get_adapter_for_lang', lambda lang: MockAdapter())
    
    results = Validator.validate(content)
    assert len(results["errors"]) == 0
    assert any("passed validation" in p for p in results["passes"])

def test_validate_code_block_delegation_failure(monkeypatch):
    content = "---\ntitle: test\n---\n```python\nprint('hello')\n```"
    
    # Mock the PythonAdapter returning an error
    class MockAdapter:
        def validate_content(self, code):
            return ["Syntax Error"]
            
    monkeypatch.setattr(Validator, 'get_adapter_for_lang', lambda lang: MockAdapter())
    
    results = Validator.validate(content)
    assert any("Code block 1 (python) error: Syntax Error" in err for err in results["errors"])

def test_unknown_language_ignored():
    content = "---\ntitle: test\n---\n```unknown\nfoo\n```"
    results = Validator.validate(content)
    # Should ignore unknown language blocks
    assert len(results["errors"]) == 0
