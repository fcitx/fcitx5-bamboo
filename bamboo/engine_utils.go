/*
 * SPDX-FileCopyrightText: 2018 Luong Thanh Lam <ltlam93@gmail.com>
 * SPDX-FileCopyrightText: 2022-2022 CSSlayer <wengxt@gmail.com>
 *
 * SPDX-License-Identifier: LGPL-2.1-or-later
 *
 */

package main

import (
	"bamboo-core"
	"unicode/utf8"
)

func (e *FcitxBambooEngine) toUpper(keyRune rune) rune {
	var keyMapping = map[rune]rune{
		'[': '{',
		']': '}',
		'{': '[',
		'}': ']',
	}

	if upperSpecialKey, found := keyMapping[keyRune]; found && inKeyList(e.preeditor.GetInputMethod().AppendingKeys, keyRune) {
		keyRune = upperSpecialKey
	}
	return keyRune
}

func (e *FcitxBambooEngine) updateLastKeyWithShift(keyVal, state uint32) {
	if e.preeditor.CanProcessKey(rune(keyVal)) {
		e.lastKeyWithShift = state&FcitxShiftMask != 0
	} else {
		e.lastKeyWithShift = false
	}
}

func (e *FcitxBambooEngine) getRawKeyLen() int {
	return len(e.preeditor.GetProcessedString(bamboo.EnglishMode | bamboo.FullText))
}

func (e *FcitxBambooEngine) runeCount() int {
	return utf8.RuneCountInString(e.getPreeditString())
}

func isValidState(state uint32) bool {
	if state&FcitxControlMask != 0 ||
		state&FcitxMod1Mask != 0 ||
		state&FcitxIgnoredMask != 0 ||
		state&FcitxSuperMask != 0 ||
		state&FcitxHyperMask != 0 ||
		state&FcitxMetaMask != 0 {
		return false
	}
	return true
}

func (e *FcitxBambooEngine) getCommitText(keyVal, state uint32) (string, bool) {
	var keyRune = rune(keyVal)
	oldText := e.getPreeditString()
	// restore key strokes by pressing Shift + Space
	if e.shouldRestoreKeyStrokes {
		e.shouldRestoreKeyStrokes = false
		e.preeditor.RestoreLastWord(!bamboo.HasAnyVietnameseRune(oldText))
		return e.getPreeditString(), false
	}
	if e.preeditor.CanProcessKey(keyRune) {
		if state&FcitxLockMask != 0 {
			keyRune = e.toUpper(keyRune)
		}
		e.preeditor.ProcessKey(keyRune, e.getBambooInputMode())
		if inKeyList(e.preeditor.GetInputMethod().AppendingKeys, keyRune) {
			var newText string
			if e.shouldFallbackToEnglish(true) {
				newText = e.getProcessedString(bamboo.EnglishMode)
			} else {
				newText = e.getProcessedString(bamboo.VietnameseMode)
			}
			if fullSeq := e.preeditor.GetProcessedString(bamboo.VietnameseMode); len(fullSeq) > 0 && rune(fullSeq[len(fullSeq)-1]) == keyRune {
				// [[ => [
				var ret = e.getPreeditString()
				var lastRune = rune(ret[len(ret)-1])
				var isWordBreakRune = bamboo.IsWordBreakSymbol(lastRune)
				// TODO: THIS IS HACKING
				if isWordBreakRune {
					e.preeditor.RemoveLastChar(false)
					e.preeditor.ProcessKey(' ', bamboo.EnglishMode)
				}
				return ret, isWordBreakRune
			} else if l := []rune(newText); len(l) > 0 && keyRune == l[len(l)-1] {
				// f] => f]
				var isWordBreakRune = bamboo.IsWordBreakSymbol(keyRune)
				if isWordBreakRune {
					e.preeditor.RemoveLastChar(false)
					e.preeditor.ProcessKey(' ', bamboo.EnglishMode)
				}
				return oldText + string(keyRune), isWordBreakRune
			} else {
				// ] => o?
				return e.getPreeditString(), false
			}
		} else if e.macroEnabled {
			return e.getProcessedString(bamboo.PunctuationMode), false
		} else {
			return e.getPreeditString(), false
		}
	} else if bamboo.IsWordBreakSymbol(keyRune) {
		// macro processing
		if e.macroEnabled {
			var keyS = string(keyRune)
			if keyVal == FcitxSpace && e.macroTable.HasKey(oldText) {
				e.preeditor.Reset()
				return e.expandMacro(oldText) + keyS, keyVal == FcitxSpace
			} else {
				e.preeditor.ProcessKey(keyRune, e.getBambooInputMode())
				return oldText + keyS, keyVal == FcitxSpace
			}
		}
		if bamboo.HasAnyVietnameseRune(oldText) && e.mustFallbackToEnglish() {
			e.preeditor.RestoreLastWord(false)
			newText := e.preeditor.GetProcessedString(bamboo.EnglishMode) + string(keyRune)
			e.preeditor.ProcessKey(keyRune, bamboo.EnglishMode)
			return newText, true
		}
		e.preeditor.ProcessKey(keyRune, bamboo.EnglishMode)
		return oldText + string(keyRune), true
	}
	return "", true
}

func (e *FcitxBambooEngine) getMacroText() (bool, string) {
	if !e.macroEnabled {
		return false, ""
	}
	var text = e.preeditor.GetProcessedString(bamboo.PunctuationMode)
	if e.macroTable.HasKey(text) {
		return true, e.expandMacro(text)
	}
	return false, ""
}
