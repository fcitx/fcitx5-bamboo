/*
 * SPDX-FileCopyrightText: 2026 fcitx5-bamboo contributors
 *
 * SPDX-License-Identifier: LGPL-2.1-or-later
 */
package main

import (
	"bamboo-core"
	"testing"
)

func newTestEngine() *FcitxBambooEngine {
	return &FcitxBambooEngine{
		preeditor:        bamboo.NewEngine(bamboo.ParseInputMethod(bamboo.InputMethodDefinitions, "Telex"), bamboo.EstdFlags),
		macroTable:       &MacroTable{mTable: map[string]string{}},
		autoNonVnRestore: false,
		outputCharset:    "Unicode",
	}
}

func TestCommitUsesConfiguredCharset(t *testing.T) {
	engine := newTestEngine()
	engine.outputCharset = "VNI Windows"

	engine.commitPreeditAndReset("gõ")

	want := bamboo.Encode("VNI Windows", "gõ")
	if want == "gõ" {
		t.Fatal("test input is not encoded by VNI Windows")
	}
	if engine.commitText != want {
		t.Fatalf("commitText = %q, want %q", engine.commitText, want)
	}
}

func TestShortcutModifiersAreNotPrintable(t *testing.T) {
	for _, state := range []uint32{
		FcitxControlMask,
		FcitxMod1Mask,
		FcitxHyperMask,
		FcitxSuperMask,
		FcitxMod5Mask,
		FcitxSuper2Mask,
		FcitxHyper2Mask,
		FcitxMetaMask,
	} {
		if isValidState(state) {
			t.Errorf("modifier state %#x accepted as printable", state)
		}
	}
}

func TestRestoreKeyStrokesIsImmediate(t *testing.T) {
	engine := newTestEngine()
	for _, key := range "gox" {
		engine.preeditProcessKeyEvent(uint32(key), 0)
	}
	if engine.preeditText != "gõ" {
		t.Fatalf("preedit before restore = %q, want %q", engine.preeditText, "gõ")
	}

	engine.restoreKeyStrokes()
	if engine.preeditText != "gox" {
		t.Fatalf("preedit after restore = %q, want %q", engine.preeditText, "gox")
	}
	engine.preeditProcessKeyEvent('a', 0)
	if engine.preeditText != "goxa" {
		t.Fatalf("next key was not processed: preedit = %q, want %q", engine.preeditText, "goxa")
	}
}

func TestMacroLookupIgnoresCase(t *testing.T) {
	table := &MacroTable{mTable: map[string]string{"vn": "Việt Nam"}}
	if !table.HasPrefix("V") || !table.HasKey("VN") {
		t.Fatal("uppercase input did not match lowercase macro")
	}
}

func TestEmptyMacroCaseIsSafe(t *testing.T) {
	if got := determineMacroCase(""); got != VnCaseNoChange {
		t.Fatalf("determineMacroCase(\"\") = %d, want %d", got, VnCaseNoChange)
	}
}
