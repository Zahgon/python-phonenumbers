// Copyright (C) 2011 The Libphonenumber Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Unicode utility functions.

package phonenumbers

import (
	"fmt"
	"sort"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
)

// Category is the general category of a Unicode character.
type Category string

const (
	// CategoryLetter identifies the corresponding Unicode general category.
	CategoryLetter Category = "L"
	// CategoryUppercaseLetter identifies the corresponding Unicode general category.
	CategoryUppercaseLetter Category = "Lu"
	// CategoryLowercaseLetter identifies the corresponding Unicode general category.
	CategoryLowercaseLetter Category = "Ll"
	// CategoryTitlecaseLetter identifies the corresponding Unicode general category.
	CategoryTitlecaseLetter Category = "Lt"
	// CategoryModifierLetter identifies the corresponding Unicode general category.
	CategoryModifierLetter Category = "Lm"
	// CategoryOtherLetter identifies the corresponding Unicode general category.
	CategoryOtherLetter Category = "Lo"
	// CategoryMark identifies the corresponding Unicode general category.
	CategoryMark Category = "M"
	// CategoryNonSpacingMark identifies the corresponding Unicode general category.
	CategoryNonSpacingMark Category = "Mn"
	// CategorySpacingCombiningMark identifies the corresponding Unicode general category.
	CategorySpacingCombiningMark Category = "Mc"
	// CategoryEnclosingMark identifies the corresponding Unicode general category.
	CategoryEnclosingMark Category = "Me"
	// CategoryNumber identifies the corresponding Unicode general category.
	CategoryNumber Category = "N"
	// CategoryDecimalDigitNumber identifies the corresponding Unicode general category.
	CategoryDecimalDigitNumber Category = "Nd"
	// CategoryLetterNumber identifies the corresponding Unicode general category.
	CategoryLetterNumber Category = "Nl"
	// CategoryOtherNumber identifies the corresponding Unicode general category.
	CategoryOtherNumber Category = "No"
	// CategorySymbol identifies the corresponding Unicode general category.
	CategorySymbol Category = "S"
	// CategoryMathSymbol identifies the corresponding Unicode general category.
	CategoryMathSymbol Category = "Sm"
	// CategoryCurrencySymbol identifies the corresponding Unicode general category.
	CategoryCurrencySymbol Category = "Sc"
	// CategoryModifierSymbol identifies the corresponding Unicode general category.
	CategoryModifierSymbol Category = "Sk"
	// CategoryOtherSymbol identifies the corresponding Unicode general category.
	CategoryOtherSymbol Category = "So"
	// CategoryPunctuation identifies the corresponding Unicode general category.
	CategoryPunctuation Category = "P"
	// CategoryConnectorPunctuation identifies the corresponding Unicode general category.
	CategoryConnectorPunctuation Category = "Pc"
	// CategoryDashPunctuation identifies the corresponding Unicode general category.
	CategoryDashPunctuation Category = "Pd"
	// CategoryOpenPunctuation identifies the corresponding Unicode general category.
	CategoryOpenPunctuation Category = "Ps"
	// CategoryClosePunctuation identifies the corresponding Unicode general category.
	CategoryClosePunctuation Category = "Pe"
	// CategoryInitialPunctuation identifies the corresponding Unicode general category.
	CategoryInitialPunctuation Category = "Pi"
	// CategoryFinalPunctuation identifies the corresponding Unicode general category.
	CategoryFinalPunctuation Category = "Pf"
	// CategoryOtherPunctuation identifies the corresponding Unicode general category.
	CategoryOtherPunctuation Category = "Po"
	// CategorySeparator identifies the corresponding Unicode general category.
	CategorySeparator Category = "Z"
	// CategorySpaceSeparator identifies the corresponding Unicode general category.
	CategorySpaceSeparator Category = "Zs"
	// CategoryLineSeparator identifies the corresponding Unicode general category.
	CategoryLineSeparator Category = "Zl"
	// CategoryParagraphSeparator identifies the corresponding Unicode general category.
	CategoryParagraphSeparator Category = "Zp"
	// CategoryOther identifies the corresponding Unicode general category.
	CategoryOther Category = "C"
	// CategoryControl identifies the corresponding Unicode general category.
	CategoryControl Category = "Cc"
	// CategoryFormat identifies the corresponding Unicode general category.
	CategoryFormat Category = "Cf"
	// CategorySurrogate identifies the corresponding Unicode general category.
	CategorySurrogate Category = "Cs"
	// CategoryPrivateUse identifies the corresponding Unicode general category.
	CategoryPrivateUse Category = "Co"
	// CategoryNotAssigned identifies the corresponding Unicode general category.
	CategoryNotAssigned Category = "Cn"
)

// CategoryGet returns the Unicode general-category code for uniChar.
func CategoryGet(uniChar rune) Category {
	return Category(pyrt.Category(uniChar))
}

// IsLetter determines whether uniChar is a Unicode letter.
func IsLetter(uniChar rune) bool {
	switch CategoryGet(uniChar) {
	case CategoryUppercaseLetter, CategoryLowercaseLetter, CategoryTitlecaseLetter,
		CategoryModifierLetter, CategoryOtherLetter:
		return true
	default:
		return false
	}
}

type blockRange struct {
	start rune
	end   rune
}

// String returns the human-readable representation.
func (r blockRange) String() string {
	return fmt.Sprintf("Block[%04x, %04x]", r.start, r.end)
}

var (
	// BlockBasicLatin identifies the corresponding Unicode block.
	BlockBasicLatin = blockRange{start: 0x0000, end: 0x007F}
	// BlockLatin1Supplement identifies the corresponding Unicode block.
	BlockLatin1Supplement = blockRange{start: 0x0080, end: 0x00FF}
	// BlockLatinExtendedA identifies the corresponding Unicode block.
	BlockLatinExtendedA = blockRange{start: 0x0100, end: 0x017F}
	// BlockLatinExtendedB identifies the corresponding Unicode block.
	BlockLatinExtendedB = blockRange{start: 0x0180, end: 0x024F}
	// BlockIpaExtensions identifies the corresponding Unicode block.
	BlockIpaExtensions = blockRange{start: 0x0250, end: 0x02AF}
	// BlockSpacingModifierLetters identifies the corresponding Unicode block.
	BlockSpacingModifierLetters = blockRange{start: 0x02B0, end: 0x02FF}
	// BlockCombiningDiacriticalMarks identifies the corresponding Unicode block.
	BlockCombiningDiacriticalMarks = blockRange{start: 0x0300, end: 0x036F}
	// BlockGreekAndCoptic identifies the corresponding Unicode block.
	BlockGreekAndCoptic = blockRange{start: 0x0370, end: 0x03FF}
	// BlockCyrillic identifies the corresponding Unicode block.
	BlockCyrillic = blockRange{start: 0x0400, end: 0x04FF}
	// BlockCyrillicSupplement identifies the corresponding Unicode block.
	BlockCyrillicSupplement = blockRange{start: 0x0500, end: 0x052F}
	// BlockArmenian identifies the corresponding Unicode block.
	BlockArmenian = blockRange{start: 0x0530, end: 0x058F}
	// BlockHebrew identifies the corresponding Unicode block.
	BlockHebrew = blockRange{start: 0x0590, end: 0x05FF}
	// BlockArabic identifies the corresponding Unicode block.
	BlockArabic = blockRange{start: 0x0600, end: 0x06FF}
	// BlockSyriac identifies the corresponding Unicode block.
	BlockSyriac = blockRange{start: 0x0700, end: 0x074F}
	// BlockArabicSupplement identifies the corresponding Unicode block.
	BlockArabicSupplement = blockRange{start: 0x0750, end: 0x077F}
	// BlockThaana identifies the corresponding Unicode block.
	BlockThaana = blockRange{start: 0x0780, end: 0x07BF}
	// BlockNko identifies the corresponding Unicode block.
	BlockNko = blockRange{start: 0x07C0, end: 0x07FF}
	// BlockSamaritan identifies the corresponding Unicode block.
	BlockSamaritan = blockRange{start: 0x0800, end: 0x083F}
	// BlockMandaic identifies the corresponding Unicode block.
	BlockMandaic = blockRange{start: 0x0840, end: 0x085F}
	// BlockDevanagari identifies the corresponding Unicode block.
	BlockDevanagari = blockRange{start: 0x0900, end: 0x097F}
	// BlockBengali identifies the corresponding Unicode block.
	BlockBengali = blockRange{start: 0x0980, end: 0x09FF}
	// BlockGurmukhi identifies the corresponding Unicode block.
	BlockGurmukhi = blockRange{start: 0x0A00, end: 0x0A7F}
	// BlockGujarati identifies the corresponding Unicode block.
	BlockGujarati = blockRange{start: 0x0A80, end: 0x0AFF}
	// BlockOriya identifies the corresponding Unicode block.
	BlockOriya = blockRange{start: 0x0B00, end: 0x0B7F}
	// BlockTamil identifies the corresponding Unicode block.
	BlockTamil = blockRange{start: 0x0B80, end: 0x0BFF}
	// BlockTelugu identifies the corresponding Unicode block.
	BlockTelugu = blockRange{start: 0x0C00, end: 0x0C7F}
	// BlockKannada identifies the corresponding Unicode block.
	BlockKannada = blockRange{start: 0x0C80, end: 0x0CFF}
	// BlockMalayalam identifies the corresponding Unicode block.
	BlockMalayalam = blockRange{start: 0x0D00, end: 0x0D7F}
	// BlockSinhala identifies the corresponding Unicode block.
	BlockSinhala = blockRange{start: 0x0D80, end: 0x0DFF}
	// BlockThai identifies the corresponding Unicode block.
	BlockThai = blockRange{start: 0x0E00, end: 0x0E7F}
	// BlockLao identifies the corresponding Unicode block.
	BlockLao = blockRange{start: 0x0E80, end: 0x0EFF}
	// BlockTibetan identifies the corresponding Unicode block.
	BlockTibetan = blockRange{start: 0x0F00, end: 0x0FFF}
	// BlockMyanmar identifies the corresponding Unicode block.
	BlockMyanmar = blockRange{start: 0x1000, end: 0x109F}
	// BlockGeorgian identifies the corresponding Unicode block.
	BlockGeorgian = blockRange{start: 0x10A0, end: 0x10FF}
	// BlockHangulJamo identifies the corresponding Unicode block.
	BlockHangulJamo = blockRange{start: 0x1100, end: 0x11FF}
	// BlockEthiopic identifies the corresponding Unicode block.
	BlockEthiopic = blockRange{start: 0x1200, end: 0x137F}
	// BlockEthiopicSupplement identifies the corresponding Unicode block.
	BlockEthiopicSupplement = blockRange{start: 0x1380, end: 0x139F}
	// BlockCherokee identifies the corresponding Unicode block.
	BlockCherokee = blockRange{start: 0x13A0, end: 0x13FF}
	// BlockUnifiedCanadianAboriginalSyllabics identifies the corresponding Unicode block.
	BlockUnifiedCanadianAboriginalSyllabics = blockRange{start: 0x1400, end: 0x167F}
	// BlockOgham identifies the corresponding Unicode block.
	BlockOgham = blockRange{start: 0x1680, end: 0x169F}
	// BlockRunic identifies the corresponding Unicode block.
	BlockRunic = blockRange{start: 0x16A0, end: 0x16FF}
	// BlockTagalog identifies the corresponding Unicode block.
	BlockTagalog = blockRange{start: 0x1700, end: 0x171F}
	// BlockHanunoo identifies the corresponding Unicode block.
	BlockHanunoo = blockRange{start: 0x1720, end: 0x173F}
	// BlockBuhid identifies the corresponding Unicode block.
	BlockBuhid = blockRange{start: 0x1740, end: 0x175F}
	// BlockTagbanwa identifies the corresponding Unicode block.
	BlockTagbanwa = blockRange{start: 0x1760, end: 0x177F}
	// BlockKhmer identifies the corresponding Unicode block.
	BlockKhmer = blockRange{start: 0x1780, end: 0x17FF}
	// BlockMongolian identifies the corresponding Unicode block.
	BlockMongolian = blockRange{start: 0x1800, end: 0x18AF}
	// BlockUnifiedCanadianAboriginalSyllabicsExtended identifies the corresponding Unicode block.
	BlockUnifiedCanadianAboriginalSyllabicsExtended = blockRange{start: 0x18B0, end: 0x18FF}
	// BlockLimbu identifies the corresponding Unicode block.
	BlockLimbu = blockRange{start: 0x1900, end: 0x194F}
	// BlockTaiLe identifies the corresponding Unicode block.
	BlockTaiLe = blockRange{start: 0x1950, end: 0x197F}
	// BlockNewTaiLue identifies the corresponding Unicode block.
	BlockNewTaiLue = blockRange{start: 0x1980, end: 0x19DF}
	// BlockKhmerSymbols identifies the corresponding Unicode block.
	BlockKhmerSymbols = blockRange{start: 0x19E0, end: 0x19FF}
	// BlockBuginese identifies the corresponding Unicode block.
	BlockBuginese = blockRange{start: 0x1A00, end: 0x1A1F}
	// BlockTaiTham identifies the corresponding Unicode block.
	BlockTaiTham = blockRange{start: 0x1A20, end: 0x1AAF}
	// BlockBalinese identifies the corresponding Unicode block.
	BlockBalinese = blockRange{start: 0x1B00, end: 0x1B7F}
	// BlockSundanese identifies the corresponding Unicode block.
	BlockSundanese = blockRange{start: 0x1B80, end: 0x1BBF}
	// BlockBatak identifies the corresponding Unicode block.
	BlockBatak = blockRange{start: 0x1BC0, end: 0x1BFF}
	// BlockLepcha identifies the corresponding Unicode block.
	BlockLepcha = blockRange{start: 0x1C00, end: 0x1C4F}
	// BlockOlChiki identifies the corresponding Unicode block.
	BlockOlChiki = blockRange{start: 0x1C50, end: 0x1C7F}
	// BlockVedicExtensions identifies the corresponding Unicode block.
	BlockVedicExtensions = blockRange{start: 0x1CD0, end: 0x1CFF}
	// BlockPhoneticExtensions identifies the corresponding Unicode block.
	BlockPhoneticExtensions = blockRange{start: 0x1D00, end: 0x1D7F}
	// BlockPhoneticExtensionsSupplement identifies the corresponding Unicode block.
	BlockPhoneticExtensionsSupplement = blockRange{start: 0x1D80, end: 0x1DBF}
	// BlockCombiningDiacriticalMarksSupplement identifies the corresponding Unicode block.
	BlockCombiningDiacriticalMarksSupplement = blockRange{start: 0x1DC0, end: 0x1DFF}
	// BlockLatinExtendedAdditional identifies the corresponding Unicode block.
	BlockLatinExtendedAdditional = blockRange{start: 0x1E00, end: 0x1EFF}
	// BlockGreekExtended identifies the corresponding Unicode block.
	BlockGreekExtended = blockRange{start: 0x1F00, end: 0x1FFF}
	// BlockGeneralPunctuation identifies the corresponding Unicode block.
	BlockGeneralPunctuation = blockRange{start: 0x2000, end: 0x206F}
	// BlockSuperscriptsAndSubscripts identifies the corresponding Unicode block.
	BlockSuperscriptsAndSubscripts = blockRange{start: 0x2070, end: 0x209F}
	// BlockCurrencySymbols identifies the corresponding Unicode block.
	BlockCurrencySymbols = blockRange{start: 0x20A0, end: 0x20CF}
	// BlockCombiningDiacriticalMarksForSymbols identifies the corresponding Unicode block.
	BlockCombiningDiacriticalMarksForSymbols = blockRange{start: 0x20D0, end: 0x20FF}
	// BlockLetterlikeSymbols identifies the corresponding Unicode block.
	BlockLetterlikeSymbols = blockRange{start: 0x2100, end: 0x214F}
	// BlockNumberForms identifies the corresponding Unicode block.
	BlockNumberForms = blockRange{start: 0x2150, end: 0x218F}
	// BlockArrows identifies the corresponding Unicode block.
	BlockArrows = blockRange{start: 0x2190, end: 0x21FF}
	// BlockMathematicalOperators identifies the corresponding Unicode block.
	BlockMathematicalOperators = blockRange{start: 0x2200, end: 0x22FF}
	// BlockMiscellaneousTechnical identifies the corresponding Unicode block.
	BlockMiscellaneousTechnical = blockRange{start: 0x2300, end: 0x23FF}
	// BlockControlPictures identifies the corresponding Unicode block.
	BlockControlPictures = blockRange{start: 0x2400, end: 0x243F}
	// BlockOpticalCharacterRecognition identifies the corresponding Unicode block.
	BlockOpticalCharacterRecognition = blockRange{start: 0x2440, end: 0x245F}
	// BlockEnclosedAlphanumerics identifies the corresponding Unicode block.
	BlockEnclosedAlphanumerics = blockRange{start: 0x2460, end: 0x24FF}
	// BlockBoxDrawing identifies the corresponding Unicode block.
	BlockBoxDrawing = blockRange{start: 0x2500, end: 0x257F}
	// BlockBlockElements identifies the corresponding Unicode block.
	BlockBlockElements = blockRange{start: 0x2580, end: 0x259F}
	// BlockGeometricShapes identifies the corresponding Unicode block.
	BlockGeometricShapes = blockRange{start: 0x25A0, end: 0x25FF}
	// BlockMiscellaneousSymbols identifies the corresponding Unicode block.
	BlockMiscellaneousSymbols = blockRange{start: 0x2600, end: 0x26FF}
	// BlockDingbats identifies the corresponding Unicode block.
	BlockDingbats = blockRange{start: 0x2700, end: 0x27BF}
	// BlockMiscellaneousMathematicalSymbolsA identifies the corresponding Unicode block.
	BlockMiscellaneousMathematicalSymbolsA = blockRange{start: 0x27C0, end: 0x27EF}
	// BlockSupplementalArrowsA identifies the corresponding Unicode block.
	BlockSupplementalArrowsA = blockRange{start: 0x27F0, end: 0x27FF}
	// BlockBraillePatterns identifies the corresponding Unicode block.
	BlockBraillePatterns = blockRange{start: 0x2800, end: 0x28FF}
	// BlockSupplementalArrowsB identifies the corresponding Unicode block.
	BlockSupplementalArrowsB = blockRange{start: 0x2900, end: 0x297F}
	// BlockMiscellaneousMathematicalSymbolsB identifies the corresponding Unicode block.
	BlockMiscellaneousMathematicalSymbolsB = blockRange{start: 0x2980, end: 0x29FF}
	// BlockSupplementalMathematicalOperators identifies the corresponding Unicode block.
	BlockSupplementalMathematicalOperators = blockRange{start: 0x2A00, end: 0x2AFF}
	// BlockMiscellaneousSymbolsAndArrows identifies the corresponding Unicode block.
	BlockMiscellaneousSymbolsAndArrows = blockRange{start: 0x2B00, end: 0x2BFF}
	// BlockGlagolitic identifies the corresponding Unicode block.
	BlockGlagolitic = blockRange{start: 0x2C00, end: 0x2C5F}
	// BlockLatinExtendedC identifies the corresponding Unicode block.
	BlockLatinExtendedC = blockRange{start: 0x2C60, end: 0x2C7F}
	// BlockCoptic identifies the corresponding Unicode block.
	BlockCoptic = blockRange{start: 0x2C80, end: 0x2CFF}
	// BlockGeorgianSupplement identifies the corresponding Unicode block.
	BlockGeorgianSupplement = blockRange{start: 0x2D00, end: 0x2D2F}
	// BlockTifinagh identifies the corresponding Unicode block.
	BlockTifinagh = blockRange{start: 0x2D30, end: 0x2D7F}
	// BlockEthiopicExtended identifies the corresponding Unicode block.
	BlockEthiopicExtended = blockRange{start: 0x2D80, end: 0x2DDF}
	// BlockCyrillicExtendedA identifies the corresponding Unicode block.
	BlockCyrillicExtendedA = blockRange{start: 0x2DE0, end: 0x2DFF}
	// BlockSupplementalPunctuation identifies the corresponding Unicode block.
	BlockSupplementalPunctuation = blockRange{start: 0x2E00, end: 0x2E7F}
	// BlockCjkRadicalsSupplement identifies the corresponding Unicode block.
	BlockCjkRadicalsSupplement = blockRange{start: 0x2E80, end: 0x2EFF}
	// BlockKangxiRadicals identifies the corresponding Unicode block.
	BlockKangxiRadicals = blockRange{start: 0x2F00, end: 0x2FDF}
	// BlockIdeographicDescriptionCharacters identifies the corresponding Unicode block.
	BlockIdeographicDescriptionCharacters = blockRange{start: 0x2FF0, end: 0x2FFF}
	// BlockCjkSymbolsAndPunctuation identifies the corresponding Unicode block.
	BlockCjkSymbolsAndPunctuation = blockRange{start: 0x3000, end: 0x303F}
	// BlockHiragana identifies the corresponding Unicode block.
	BlockHiragana = blockRange{start: 0x3040, end: 0x309F}
	// BlockKatakana identifies the corresponding Unicode block.
	BlockKatakana = blockRange{start: 0x30A0, end: 0x30FF}
	// BlockBopomofo identifies the corresponding Unicode block.
	BlockBopomofo = blockRange{start: 0x3100, end: 0x312F}
	// BlockHangulCompatibilityJamo identifies the corresponding Unicode block.
	BlockHangulCompatibilityJamo = blockRange{start: 0x3130, end: 0x318F}
	// BlockKanbun identifies the corresponding Unicode block.
	BlockKanbun = blockRange{start: 0x3190, end: 0x319F}
	// BlockBopomofoExtended identifies the corresponding Unicode block.
	BlockBopomofoExtended = blockRange{start: 0x31A0, end: 0x31BF}
	// BlockCjkStrokes identifies the corresponding Unicode block.
	BlockCjkStrokes = blockRange{start: 0x31C0, end: 0x31EF}
	// BlockKatakanaPhoneticExtensions identifies the corresponding Unicode block.
	BlockKatakanaPhoneticExtensions = blockRange{start: 0x31F0, end: 0x31FF}
	// BlockEnclosedCjkLettersAndMonths identifies the corresponding Unicode block.
	BlockEnclosedCjkLettersAndMonths = blockRange{start: 0x3200, end: 0x32FF}
	// BlockCjkCompatibility identifies the corresponding Unicode block.
	BlockCjkCompatibility = blockRange{start: 0x3300, end: 0x33FF}
	// BlockCjkUnifiedIdeographsExtensionA identifies the corresponding Unicode block.
	BlockCjkUnifiedIdeographsExtensionA = blockRange{start: 0x3400, end: 0x4DBF}
	// BlockYijingHexagramSymbols identifies the corresponding Unicode block.
	BlockYijingHexagramSymbols = blockRange{start: 0x4DC0, end: 0x4DFF}
	// BlockCjkUnifiedIdeographs identifies the corresponding Unicode block.
	BlockCjkUnifiedIdeographs = blockRange{start: 0x4E00, end: 0x9FFF}
	// BlockYiSyllables identifies the corresponding Unicode block.
	BlockYiSyllables = blockRange{start: 0xA000, end: 0xA48F}
	// BlockYiRadicals identifies the corresponding Unicode block.
	BlockYiRadicals = blockRange{start: 0xA490, end: 0xA4CF}
	// BlockLisu identifies the corresponding Unicode block.
	BlockLisu = blockRange{start: 0xA4D0, end: 0xA4FF}
	// BlockVai identifies the corresponding Unicode block.
	BlockVai = blockRange{start: 0xA500, end: 0xA63F}
	// BlockCyrillicExtendedB identifies the corresponding Unicode block.
	BlockCyrillicExtendedB = blockRange{start: 0xA640, end: 0xA69F}
	// BlockBamum identifies the corresponding Unicode block.
	BlockBamum = blockRange{start: 0xA6A0, end: 0xA6FF}
	// BlockModifierToneLetters identifies the corresponding Unicode block.
	BlockModifierToneLetters = blockRange{start: 0xA700, end: 0xA71F}
	// BlockLatinExtendedD identifies the corresponding Unicode block.
	BlockLatinExtendedD = blockRange{start: 0xA720, end: 0xA7FF}
	// BlockSylotiNagri identifies the corresponding Unicode block.
	BlockSylotiNagri = blockRange{start: 0xA800, end: 0xA82F}
	// BlockCommonIndicNumberForms identifies the corresponding Unicode block.
	BlockCommonIndicNumberForms = blockRange{start: 0xA830, end: 0xA83F}
	// BlockPhagsPa identifies the corresponding Unicode block.
	BlockPhagsPa = blockRange{start: 0xA840, end: 0xA87F}
	// BlockSaurashtra identifies the corresponding Unicode block.
	BlockSaurashtra = blockRange{start: 0xA880, end: 0xA8DF}
	// BlockDevanagariExtended identifies the corresponding Unicode block.
	BlockDevanagariExtended = blockRange{start: 0xA8E0, end: 0xA8FF}
	// BlockKayahLi identifies the corresponding Unicode block.
	BlockKayahLi = blockRange{start: 0xA900, end: 0xA92F}
	// BlockRejang identifies the corresponding Unicode block.
	BlockRejang = blockRange{start: 0xA930, end: 0xA95F}
	// BlockHangulJamoExtendedA identifies the corresponding Unicode block.
	BlockHangulJamoExtendedA = blockRange{start: 0xA960, end: 0xA97F}
	// BlockJavanese identifies the corresponding Unicode block.
	BlockJavanese = blockRange{start: 0xA980, end: 0xA9DF}
	// BlockCham identifies the corresponding Unicode block.
	BlockCham = blockRange{start: 0xAA00, end: 0xAA5F}
	// BlockMyanmarExtendedA identifies the corresponding Unicode block.
	BlockMyanmarExtendedA = blockRange{start: 0xAA60, end: 0xAA7F}
	// BlockTaiViet identifies the corresponding Unicode block.
	BlockTaiViet = blockRange{start: 0xAA80, end: 0xAADF}
	// BlockEthiopicExtendedA identifies the corresponding Unicode block.
	BlockEthiopicExtendedA = blockRange{start: 0xAB00, end: 0xAB2F}
	// BlockMeeteiMayek identifies the corresponding Unicode block.
	BlockMeeteiMayek = blockRange{start: 0xABC0, end: 0xABFF}
	// BlockHangulSyllables identifies the corresponding Unicode block.
	BlockHangulSyllables = blockRange{start: 0xAC00, end: 0xD7AF}
	// BlockHangulJamoExtendedB identifies the corresponding Unicode block.
	BlockHangulJamoExtendedB = blockRange{start: 0xD7B0, end: 0xD7FF}
	// BlockHighSurrogates identifies the corresponding Unicode block.
	BlockHighSurrogates = blockRange{start: 0xD800, end: 0xDB7F}
	// BlockHighPrivateUseSurrogates identifies the corresponding Unicode block.
	BlockHighPrivateUseSurrogates = blockRange{start: 0xDB80, end: 0xDBFF}
	// BlockLowSurrogates identifies the corresponding Unicode block.
	BlockLowSurrogates = blockRange{start: 0xDC00, end: 0xDFFF}
	// BlockPrivateUseArea identifies the corresponding Unicode block.
	BlockPrivateUseArea = blockRange{start: 0xE000, end: 0xF8FF}
	// BlockCjkCompatibilityIdeographs identifies the corresponding Unicode block.
	BlockCjkCompatibilityIdeographs = blockRange{start: 0xF900, end: 0xFAFF}
	// BlockAlphabeticPresentationForms identifies the corresponding Unicode block.
	BlockAlphabeticPresentationForms = blockRange{start: 0xFB00, end: 0xFB4F}
	// BlockArabicPresentationFormsA identifies the corresponding Unicode block.
	BlockArabicPresentationFormsA = blockRange{start: 0xFB50, end: 0xFDFF}
	// BlockVariationSelectors identifies the corresponding Unicode block.
	BlockVariationSelectors = blockRange{start: 0xFE00, end: 0xFE0F}
	// BlockVerticalForms identifies the corresponding Unicode block.
	BlockVerticalForms = blockRange{start: 0xFE10, end: 0xFE1F}
	// BlockCombiningHalfMarks identifies the corresponding Unicode block.
	BlockCombiningHalfMarks = blockRange{start: 0xFE20, end: 0xFE2F}
	// BlockCjkCompatibilityForms identifies the corresponding Unicode block.
	BlockCjkCompatibilityForms = blockRange{start: 0xFE30, end: 0xFE4F}
	// BlockSmallFormVariants identifies the corresponding Unicode block.
	BlockSmallFormVariants = blockRange{start: 0xFE50, end: 0xFE6F}
	// BlockArabicPresentationFormsB identifies the corresponding Unicode block.
	BlockArabicPresentationFormsB = blockRange{start: 0xFE70, end: 0xFEFF}
	// BlockHalfwidthAndFullwidthForms identifies the corresponding Unicode block.
	BlockHalfwidthAndFullwidthForms = blockRange{start: 0xFF00, end: 0xFFEF}
	// BlockSpecials identifies the corresponding Unicode block.
	BlockSpecials = blockRange{start: 0xFFF0, end: 0xFFFF}
	// BlockLinearBSyllabary identifies the corresponding Unicode block.
	BlockLinearBSyllabary = blockRange{start: 0x10000, end: 0x1007F}
	// BlockLinearBIdeograms identifies the corresponding Unicode block.
	BlockLinearBIdeograms = blockRange{start: 0x10080, end: 0x100FF}
	// BlockAegeanNumbers identifies the corresponding Unicode block.
	BlockAegeanNumbers = blockRange{start: 0x10100, end: 0x1013F}
	// BlockAncientGreekNumbers identifies the corresponding Unicode block.
	BlockAncientGreekNumbers = blockRange{start: 0x10140, end: 0x1018F}
	// BlockAncientSymbols identifies the corresponding Unicode block.
	BlockAncientSymbols = blockRange{start: 0x10190, end: 0x101CF}
	// BlockPhaistosDisc identifies the corresponding Unicode block.
	BlockPhaistosDisc = blockRange{start: 0x101D0, end: 0x101FF}
	// BlockLycian identifies the corresponding Unicode block.
	BlockLycian = blockRange{start: 0x10280, end: 0x1029F}
	// BlockCarian identifies the corresponding Unicode block.
	BlockCarian = blockRange{start: 0x102A0, end: 0x102DF}
	// BlockOldItalic identifies the corresponding Unicode block.
	BlockOldItalic = blockRange{start: 0x10300, end: 0x1032F}
	// BlockGothic identifies the corresponding Unicode block.
	BlockGothic = blockRange{start: 0x10330, end: 0x1034F}
	// BlockUgaritic identifies the corresponding Unicode block.
	BlockUgaritic = blockRange{start: 0x10380, end: 0x1039F}
	// BlockOldPersian identifies the corresponding Unicode block.
	BlockOldPersian = blockRange{start: 0x103A0, end: 0x103DF}
	// BlockDeseret identifies the corresponding Unicode block.
	BlockDeseret = blockRange{start: 0x10400, end: 0x1044F}
	// BlockShavian identifies the corresponding Unicode block.
	BlockShavian = blockRange{start: 0x10450, end: 0x1047F}
	// BlockOsmanya identifies the corresponding Unicode block.
	BlockOsmanya = blockRange{start: 0x10480, end: 0x104AF}
	// BlockCypriotSyllabary identifies the corresponding Unicode block.
	BlockCypriotSyllabary = blockRange{start: 0x10800, end: 0x1083F}
	// BlockImperialAramaic identifies the corresponding Unicode block.
	BlockImperialAramaic = blockRange{start: 0x10840, end: 0x1085F}
	// BlockPhoenician identifies the corresponding Unicode block.
	BlockPhoenician = blockRange{start: 0x10900, end: 0x1091F}
	// BlockLydian identifies the corresponding Unicode block.
	BlockLydian = blockRange{start: 0x10920, end: 0x1093F}
	// BlockKharoshthi identifies the corresponding Unicode block.
	BlockKharoshthi = blockRange{start: 0x10A00, end: 0x10A5F}
	// BlockOldSouthArabian identifies the corresponding Unicode block.
	BlockOldSouthArabian = blockRange{start: 0x10A60, end: 0x10A7F}
	// BlockAvestan identifies the corresponding Unicode block.
	BlockAvestan = blockRange{start: 0x10B00, end: 0x10B3F}
	// BlockInscriptionalParthian identifies the corresponding Unicode block.
	BlockInscriptionalParthian = blockRange{start: 0x10B40, end: 0x10B5F}
	// BlockInscriptionalPahlavi identifies the corresponding Unicode block.
	BlockInscriptionalPahlavi = blockRange{start: 0x10B60, end: 0x10B7F}
	// BlockOldTurkic identifies the corresponding Unicode block.
	BlockOldTurkic = blockRange{start: 0x10C00, end: 0x10C4F}
	// BlockRumiNumeralSymbols identifies the corresponding Unicode block.
	BlockRumiNumeralSymbols = blockRange{start: 0x10E60, end: 0x10E7F}
	// BlockBrahmi identifies the corresponding Unicode block.
	BlockBrahmi = blockRange{start: 0x11000, end: 0x1107F}
	// BlockKaithi identifies the corresponding Unicode block.
	BlockKaithi = blockRange{start: 0x11080, end: 0x110CF}
	// BlockCuneiform identifies the corresponding Unicode block.
	BlockCuneiform = blockRange{start: 0x12000, end: 0x123FF}
	// BlockCuneiformNumbersAndPunctuation identifies the corresponding Unicode block.
	BlockCuneiformNumbersAndPunctuation = blockRange{start: 0x12400, end: 0x1247F}
	// BlockEgyptianHieroglyphs identifies the corresponding Unicode block.
	BlockEgyptianHieroglyphs = blockRange{start: 0x13000, end: 0x1342F}
	// BlockBamumSupplement identifies the corresponding Unicode block.
	BlockBamumSupplement = blockRange{start: 0x16800, end: 0x16A3F}
	// BlockKanaSupplement identifies the corresponding Unicode block.
	BlockKanaSupplement = blockRange{start: 0x1B000, end: 0x1B0FF}
	// BlockByzantineMusicalSymbols identifies the corresponding Unicode block.
	BlockByzantineMusicalSymbols = blockRange{start: 0x1D000, end: 0x1D0FF}
	// BlockMusicalSymbols identifies the corresponding Unicode block.
	BlockMusicalSymbols = blockRange{start: 0x1D100, end: 0x1D1FF}
	// BlockAncientGreekMusicalNotation identifies the corresponding Unicode block.
	BlockAncientGreekMusicalNotation = blockRange{start: 0x1D200, end: 0x1D24F}
	// BlockTaiXuanJingSymbols identifies the corresponding Unicode block.
	BlockTaiXuanJingSymbols = blockRange{start: 0x1D300, end: 0x1D35F}
	// BlockCountingRodNumerals identifies the corresponding Unicode block.
	BlockCountingRodNumerals = blockRange{start: 0x1D360, end: 0x1D37F}
	// BlockMathematicalAlphanumericSymbols identifies the corresponding Unicode block.
	BlockMathematicalAlphanumericSymbols = blockRange{start: 0x1D400, end: 0x1D7FF}
	// BlockMahjongTiles identifies the corresponding Unicode block.
	BlockMahjongTiles = blockRange{start: 0x1F000, end: 0x1F02F}
	// BlockDominoTiles identifies the corresponding Unicode block.
	BlockDominoTiles = blockRange{start: 0x1F030, end: 0x1F09F}
	// BlockPlayingCards identifies the corresponding Unicode block.
	BlockPlayingCards = blockRange{start: 0x1F0A0, end: 0x1F0FF}
	// BlockEnclosedAlphanumericSupplement identifies the corresponding Unicode block.
	BlockEnclosedAlphanumericSupplement = blockRange{start: 0x1F100, end: 0x1F1FF}
	// BlockEnclosedIdeographicSupplement identifies the corresponding Unicode block.
	BlockEnclosedIdeographicSupplement = blockRange{start: 0x1F200, end: 0x1F2FF}
	// BlockMiscellaneousSymbolsAndPictographs identifies the corresponding Unicode block.
	BlockMiscellaneousSymbolsAndPictographs = blockRange{start: 0x1F300, end: 0x1F5FF}
	// BlockEmoticons identifies the corresponding Unicode block.
	BlockEmoticons = blockRange{start: 0x1F600, end: 0x1F64F}
	// BlockTransportAndMapSymbols identifies the corresponding Unicode block.
	BlockTransportAndMapSymbols = blockRange{start: 0x1F680, end: 0x1F6FF}
	// BlockAlchemicalSymbols identifies the corresponding Unicode block.
	BlockAlchemicalSymbols = blockRange{start: 0x1F700, end: 0x1F77F}
	// BlockCjkUnifiedIdeographsExtensionB identifies the corresponding Unicode block.
	BlockCjkUnifiedIdeographsExtensionB = blockRange{start: 0x20000, end: 0x2A6DF}
	// BlockCjkUnifiedIdeographsExtensionC identifies the corresponding Unicode block.
	BlockCjkUnifiedIdeographsExtensionC = blockRange{start: 0x2A700, end: 0x2B73F}
	// BlockCjkUnifiedIdeographsExtensionD identifies the corresponding Unicode block.
	BlockCjkUnifiedIdeographsExtensionD = blockRange{start: 0x2B740, end: 0x2B81F}
	// BlockCjkCompatibilityIdeographsSupplement identifies the corresponding Unicode block.
	BlockCjkCompatibilityIdeographsSupplement = blockRange{start: 0x2F800, end: 0x2FA1F}
	// BlockTags identifies the corresponding Unicode block.
	BlockTags = blockRange{start: 0xE0000, end: 0xE007F}
	// BlockVariationSelectorsSupplement identifies the corresponding Unicode block.
	BlockVariationSelectorsSupplement = blockRange{start: 0xE0100, end: 0xE01EF}
	// BlockSupplementaryPrivateUseAreaA identifies the corresponding Unicode block.
	BlockSupplementaryPrivateUseAreaA = blockRange{start: 0xF0000, end: 0xFFFFF}
	// BlockSupplementaryPrivateUseAreaB identifies the corresponding Unicode block.
	BlockSupplementaryPrivateUseAreaB = blockRange{start: 0x100000, end: 0x10FFFF}
	// BlockUnknown identifies the corresponding Unicode block.
	BlockUnknown = blockRange{start: -1, end: -1}
)

var blockRanges = []blockRange{
	BlockBasicLatin,
	BlockLatin1Supplement,
	BlockLatinExtendedA,
	BlockLatinExtendedB,
	BlockIpaExtensions,
	BlockSpacingModifierLetters,
	BlockCombiningDiacriticalMarks,
	BlockGreekAndCoptic,
	BlockCyrillic,
	BlockCyrillicSupplement,
	BlockArmenian,
	BlockHebrew,
	BlockArabic,
	BlockSyriac,
	BlockArabicSupplement,
	BlockThaana,
	BlockNko,
	BlockSamaritan,
	BlockMandaic,
	BlockDevanagari,
	BlockBengali,
	BlockGurmukhi,
	BlockGujarati,
	BlockOriya,
	BlockTamil,
	BlockTelugu,
	BlockKannada,
	BlockMalayalam,
	BlockSinhala,
	BlockThai,
	BlockLao,
	BlockTibetan,
	BlockMyanmar,
	BlockGeorgian,
	BlockHangulJamo,
	BlockEthiopic,
	BlockEthiopicSupplement,
	BlockCherokee,
	BlockUnifiedCanadianAboriginalSyllabics,
	BlockOgham,
	BlockRunic,
	BlockTagalog,
	BlockHanunoo,
	BlockBuhid,
	BlockTagbanwa,
	BlockKhmer,
	BlockMongolian,
	BlockUnifiedCanadianAboriginalSyllabicsExtended,
	BlockLimbu,
	BlockTaiLe,
	BlockNewTaiLue,
	BlockKhmerSymbols,
	BlockBuginese,
	BlockTaiTham,
	BlockBalinese,
	BlockSundanese,
	BlockBatak,
	BlockLepcha,
	BlockOlChiki,
	BlockVedicExtensions,
	BlockPhoneticExtensions,
	BlockPhoneticExtensionsSupplement,
	BlockCombiningDiacriticalMarksSupplement,
	BlockLatinExtendedAdditional,
	BlockGreekExtended,
	BlockGeneralPunctuation,
	BlockSuperscriptsAndSubscripts,
	BlockCurrencySymbols,
	BlockCombiningDiacriticalMarksForSymbols,
	BlockLetterlikeSymbols,
	BlockNumberForms,
	BlockArrows,
	BlockMathematicalOperators,
	BlockMiscellaneousTechnical,
	BlockControlPictures,
	BlockOpticalCharacterRecognition,
	BlockEnclosedAlphanumerics,
	BlockBoxDrawing,
	BlockBlockElements,
	BlockGeometricShapes,
	BlockMiscellaneousSymbols,
	BlockDingbats,
	BlockMiscellaneousMathematicalSymbolsA,
	BlockSupplementalArrowsA,
	BlockBraillePatterns,
	BlockSupplementalArrowsB,
	BlockMiscellaneousMathematicalSymbolsB,
	BlockSupplementalMathematicalOperators,
	BlockMiscellaneousSymbolsAndArrows,
	BlockGlagolitic,
	BlockLatinExtendedC,
	BlockCoptic,
	BlockGeorgianSupplement,
	BlockTifinagh,
	BlockEthiopicExtended,
	BlockCyrillicExtendedA,
	BlockSupplementalPunctuation,
	BlockCjkRadicalsSupplement,
	BlockKangxiRadicals,
	BlockIdeographicDescriptionCharacters,
	BlockCjkSymbolsAndPunctuation,
	BlockHiragana,
	BlockKatakana,
	BlockBopomofo,
	BlockHangulCompatibilityJamo,
	BlockKanbun,
	BlockBopomofoExtended,
	BlockCjkStrokes,
	BlockKatakanaPhoneticExtensions,
	BlockEnclosedCjkLettersAndMonths,
	BlockCjkCompatibility,
	BlockCjkUnifiedIdeographsExtensionA,
	BlockYijingHexagramSymbols,
	BlockCjkUnifiedIdeographs,
	BlockYiSyllables,
	BlockYiRadicals,
	BlockLisu,
	BlockVai,
	BlockCyrillicExtendedB,
	BlockBamum,
	BlockModifierToneLetters,
	BlockLatinExtendedD,
	BlockSylotiNagri,
	BlockCommonIndicNumberForms,
	BlockPhagsPa,
	BlockSaurashtra,
	BlockDevanagariExtended,
	BlockKayahLi,
	BlockRejang,
	BlockHangulJamoExtendedA,
	BlockJavanese,
	BlockCham,
	BlockMyanmarExtendedA,
	BlockTaiViet,
	BlockEthiopicExtendedA,
	BlockMeeteiMayek,
	BlockHangulSyllables,
	BlockHangulJamoExtendedB,
	BlockHighSurrogates,
	BlockHighPrivateUseSurrogates,
	BlockLowSurrogates,
	BlockPrivateUseArea,
	BlockCjkCompatibilityIdeographs,
	BlockAlphabeticPresentationForms,
	BlockArabicPresentationFormsA,
	BlockVariationSelectors,
	BlockVerticalForms,
	BlockCombiningHalfMarks,
	BlockCjkCompatibilityForms,
	BlockSmallFormVariants,
	BlockArabicPresentationFormsB,
	BlockHalfwidthAndFullwidthForms,
	BlockSpecials,
	BlockLinearBSyllabary,
	BlockLinearBIdeograms,
	BlockAegeanNumbers,
	BlockAncientGreekNumbers,
	BlockAncientSymbols,
	BlockPhaistosDisc,
	BlockLycian,
	BlockCarian,
	BlockOldItalic,
	BlockGothic,
	BlockUgaritic,
	BlockOldPersian,
	BlockDeseret,
	BlockShavian,
	BlockOsmanya,
	BlockCypriotSyllabary,
	BlockImperialAramaic,
	BlockPhoenician,
	BlockLydian,
	BlockKharoshthi,
	BlockOldSouthArabian,
	BlockAvestan,
	BlockInscriptionalParthian,
	BlockInscriptionalPahlavi,
	BlockOldTurkic,
	BlockRumiNumeralSymbols,
	BlockBrahmi,
	BlockKaithi,
	BlockCuneiform,
	BlockCuneiformNumbersAndPunctuation,
	BlockEgyptianHieroglyphs,
	BlockBamumSupplement,
	BlockKanaSupplement,
	BlockByzantineMusicalSymbols,
	BlockMusicalSymbols,
	BlockAncientGreekMusicalNotation,
	BlockTaiXuanJingSymbols,
	BlockCountingRodNumerals,
	BlockMathematicalAlphanumericSymbols,
	BlockMahjongTiles,
	BlockDominoTiles,
	BlockPlayingCards,
	BlockEnclosedAlphanumericSupplement,
	BlockEnclosedIdeographicSupplement,
	BlockMiscellaneousSymbolsAndPictographs,
	BlockEmoticons,
	BlockTransportAndMapSymbols,
	BlockAlchemicalSymbols,
	BlockCjkUnifiedIdeographsExtensionB,
	BlockCjkUnifiedIdeographsExtensionC,
	BlockCjkUnifiedIdeographsExtensionD,
	BlockCjkCompatibilityIdeographsSupplement,
	BlockTags,
	BlockVariationSelectorsSupplement,
	BlockSupplementaryPrivateUseAreaA,
	BlockSupplementaryPrivateUseAreaB,
}

// BlockGet returns the block from this module's original Unicode block table.
func BlockGet(uniChar rune) blockRange {
	i := sort.Search(len(blockRanges), func(i int) bool {
		return blockRanges[i].start >= uniChar
	})
	if i > 0 {
		prev := blockRanges[i-1]
		if uniChar >= prev.start && uniChar <= prev.end {
			return prev
		}
	}
	if i < len(blockRanges) {
		next := blockRanges[i]
		if uniChar >= next.start && uniChar <= next.end {
			return next
		}
	}
	return BlockUnknown
}

// Digit returns the digit value assigned to uniChar. If no value exists,
// defaultValue is returned when non-nil; otherwise a ValueError-equivalent
// error is returned.
func Digit(uniChar rune, defaultValue *int) (int, error) {
	if value, ok := pyrt.Digit(uniChar); ok {
		return value, nil
	}
	if defaultValue != nil {
		return *defaultValue, nil
	}
	return 0, &NotADigitError{}
}

// digit is the core library's default-valued unicodedata.digit helper.
func digit(r rune, def int) int {
	if value, ok := pyrt.Digit(r); ok {
		return value
	}
	return def
}
