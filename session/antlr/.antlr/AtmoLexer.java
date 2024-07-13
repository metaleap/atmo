// Generated from /home/_/c/go/_/atmo/session/antlr/Atmo.g4 by ANTLR 4.9.2
import org.antlr.v4.runtime.Lexer;
import org.antlr.v4.runtime.CharStream;
import org.antlr.v4.runtime.Token;
import org.antlr.v4.runtime.TokenStream;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.misc.*;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast"})
public class AtmoLexer extends Lexer {
	static { RuntimeMetaData.checkVersion("4.9.2", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		L_PAREN=1, R_PAREN=2, L_CURLY=3, R_CURLY=4, L_BRACKET=5, R_BRACKET=6, 
		COMMA=7, NO_OP=8, IDENTIFIER=9, OPERATOR=10, DECIMAL_LIT=11, BINARY_LIT=12, 
		OCTAL_LIT=13, HEX_LIT=14, FLOAT_LIT=15, DECIMAL_FLOAT_LIT=16, HEX_FLOAT_LIT=17, 
		IMAGINARY_LIT=18, RUNE_LIT=19, BYTE_VALUE=20, OCTAL_BYTE_VALUE=21, HEX_BYTE_VALUE=22, 
		LITTLE_U_VALUE=23, BIG_U_VALUE=24, RAW_STRING_LIT=25, INTERPRETED_STRING_LIT=26, 
		COMMENT=27, LINE_COMMENT=28, WS=29, TERMINATOR=30, WS_NL=31;
	public static String[] channelNames = {
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN"
	};

	public static String[] modeNames = {
		"DEFAULT_MODE"
	};

	private static String[] makeRuleNames() {
		return new String[] {
			"L_PAREN", "R_PAREN", "L_CURLY", "R_CURLY", "L_BRACKET", "R_BRACKET", 
			"COMMA", "NO_OP", "IDENTIFIER", "OPERATOR", "DECIMAL_LIT", "BINARY_LIT", 
			"OCTAL_LIT", "HEX_LIT", "FLOAT_LIT", "DECIMAL_FLOAT_LIT", "HEX_FLOAT_LIT", 
			"HEX_MANTISSA", "HEX_EXPONENT", "IMAGINARY_LIT", "RUNE", "RUNE_LIT", 
			"BYTE_VALUE", "OCTAL_BYTE_VALUE", "HEX_BYTE_VALUE", "LITTLE_U_VALUE", 
			"BIG_U_VALUE", "RAW_STRING_LIT", "INTERPRETED_STRING_LIT", "COMMENT", 
			"LINE_COMMENT", "WS", "TERMINATOR", "WS_NL", "UNICODE_VALUE", "ESCAPED_VALUE", 
			"DECIMALS", "OCTAL_DIGIT", "HEX_DIGIT", "BIN_DIGIT", "EXPONENT", "LETTER", 
			"UNICODE_DIGIT", "UNICODE_LETTER", "UNICODE_OPISH"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "'('", "')'", "'{'", "'}'", "'['", "']'", "','"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "L_PAREN", "R_PAREN", "L_CURLY", "R_CURLY", "L_BRACKET", "R_BRACKET", 
			"COMMA", "NO_OP", "IDENTIFIER", "OPERATOR", "DECIMAL_LIT", "BINARY_LIT", 
			"OCTAL_LIT", "HEX_LIT", "FLOAT_LIT", "DECIMAL_FLOAT_LIT", "HEX_FLOAT_LIT", 
			"IMAGINARY_LIT", "RUNE_LIT", "BYTE_VALUE", "OCTAL_BYTE_VALUE", "HEX_BYTE_VALUE", 
			"LITTLE_U_VALUE", "BIG_U_VALUE", "RAW_STRING_LIT", "INTERPRETED_STRING_LIT", 
			"COMMENT", "LINE_COMMENT", "WS", "TERMINATOR", "WS_NL"
		};
	}
	private static final String[] _SYMBOLIC_NAMES = makeSymbolicNames();
	public static final Vocabulary VOCABULARY = new VocabularyImpl(_LITERAL_NAMES, _SYMBOLIC_NAMES);

	/**
	 * @deprecated Use {@link #VOCABULARY} instead.
	 */
	@Deprecated
	public static final String[] tokenNames;
	static {
		tokenNames = new String[_SYMBOLIC_NAMES.length];
		for (int i = 0; i < tokenNames.length; i++) {
			tokenNames[i] = VOCABULARY.getLiteralName(i);
			if (tokenNames[i] == null) {
				tokenNames[i] = VOCABULARY.getSymbolicName(i);
			}

			if (tokenNames[i] == null) {
				tokenNames[i] = "<INVALID>";
			}
		}
	}

	@Override
	@Deprecated
	public String[] getTokenNames() {
		return tokenNames;
	}

	@Override

	public Vocabulary getVocabulary() {
		return VOCABULARY;
	}


	public AtmoLexer(CharStream input) {
		super(input);
		_interp = new LexerATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@Override
	public String getGrammarFileName() { return "Atmo.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public String[] getChannelNames() { return channelNames; }

	@Override
	public String[] getModeNames() { return modeNames; }

	@Override
	public ATN getATN() { return _ATN; }

	public static final String _serializedATN =
		"\3\u608b\ua72a\u8133\ub9ed\u417c\u3be7\u7786\u5964\2!\u01a8\b\1\4\2\t"+
		"\2\4\3\t\3\4\4\t\4\4\5\t\5\4\6\t\6\4\7\t\7\4\b\t\b\4\t\t\t\4\n\t\n\4\13"+
		"\t\13\4\f\t\f\4\r\t\r\4\16\t\16\4\17\t\17\4\20\t\20\4\21\t\21\4\22\t\22"+
		"\4\23\t\23\4\24\t\24\4\25\t\25\4\26\t\26\4\27\t\27\4\30\t\30\4\31\t\31"+
		"\4\32\t\32\4\33\t\33\4\34\t\34\4\35\t\35\4\36\t\36\4\37\t\37\4 \t \4!"+
		"\t!\4\"\t\"\4#\t#\4$\t$\4%\t%\4&\t&\4\'\t\'\4(\t(\4)\t)\4*\t*\4+\t+\4"+
		",\t,\4-\t-\4.\t.\3\2\3\2\3\3\3\3\3\4\3\4\3\5\3\5\3\6\3\6\3\7\3\7\3\b\3"+
		"\b\3\t\3\t\3\t\3\t\3\t\3\t\3\t\5\ts\n\t\3\n\5\nv\n\n\3\n\3\n\3\n\7\n{"+
		"\n\n\f\n\16\n~\13\n\3\13\6\13\u0081\n\13\r\13\16\13\u0082\3\f\3\f\3\f"+
		"\5\f\u0088\n\f\3\f\7\f\u008b\n\f\f\f\16\f\u008e\13\f\5\f\u0090\n\f\3\r"+
		"\3\r\3\r\5\r\u0095\n\r\3\r\6\r\u0098\n\r\r\r\16\r\u0099\3\16\3\16\5\16"+
		"\u009e\n\16\3\16\5\16\u00a1\n\16\3\16\6\16\u00a4\n\16\r\16\16\16\u00a5"+
		"\3\17\3\17\3\17\5\17\u00ab\n\17\3\17\6\17\u00ae\n\17\r\17\16\17\u00af"+
		"\3\20\3\20\5\20\u00b4\n\20\3\21\3\21\3\21\5\21\u00b9\n\21\3\21\5\21\u00bc"+
		"\n\21\3\21\5\21\u00bf\n\21\3\21\3\21\3\21\5\21\u00c4\n\21\5\21\u00c6\n"+
		"\21\3\22\3\22\3\22\3\22\3\22\3\23\5\23\u00ce\n\23\3\23\6\23\u00d1\n\23"+
		"\r\23\16\23\u00d2\3\23\3\23\5\23\u00d7\n\23\3\23\7\23\u00da\n\23\f\23"+
		"\16\23\u00dd\13\23\5\23\u00df\n\23\3\23\3\23\3\23\5\23\u00e4\n\23\3\23"+
		"\7\23\u00e7\n\23\f\23\16\23\u00ea\13\23\5\23\u00ec\n\23\3\24\3\24\5\24"+
		"\u00f0\n\24\3\24\3\24\3\25\3\25\3\25\3\25\3\25\5\25\u00f9\n\25\3\25\3"+
		"\25\3\26\3\26\3\26\5\26\u0100\n\26\3\26\3\26\3\27\3\27\3\30\3\30\5\30"+
		"\u0108\n\30\3\31\3\31\3\31\3\31\3\31\3\32\3\32\3\32\3\32\3\32\3\33\3\33"+
		"\3\33\3\33\3\33\3\33\3\33\3\34\3\34\3\34\3\34\3\34\3\34\3\34\3\34\3\34"+
		"\3\34\3\34\3\35\3\35\7\35\u0128\n\35\f\35\16\35\u012b\13\35\3\35\3\35"+
		"\3\36\3\36\3\36\7\36\u0132\n\36\f\36\16\36\u0135\13\36\3\36\3\36\3\37"+
		"\3\37\3\37\3\37\7\37\u013d\n\37\f\37\16\37\u0140\13\37\3\37\3\37\3\37"+
		"\3\37\3\37\3 \3 \3 \3 \7 \u014b\n \f \16 \u014e\13 \3 \3 \3!\6!\u0153"+
		"\n!\r!\16!\u0154\3!\3!\3\"\6\"\u015a\n\"\r\"\16\"\u015b\3\"\3\"\3#\6#"+
		"\u0161\n#\r#\16#\u0162\3#\3#\3$\3$\3$\3$\5$\u016b\n$\3%\3%\3%\3%\3%\3"+
		"%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\3%\5%\u0187"+
		"\n%\3&\3&\5&\u018b\n&\3&\7&\u018e\n&\f&\16&\u0191\13&\3\'\3\'\3(\3(\3"+
		")\3)\3*\3*\5*\u019b\n*\3*\3*\3+\3+\5+\u01a1\n+\3,\3,\3-\3-\3.\3.\3\u013e"+
		"\2/\3\3\5\4\7\5\t\6\13\7\r\b\17\t\21\n\23\13\25\f\27\r\31\16\33\17\35"+
		"\20\37\21!\22#\23%\2\'\2)\24+\2-\25/\26\61\27\63\30\65\31\67\329\33;\34"+
		"=\35?\36A\37C E!G\2I\2K\2M\2O\2Q\2S\2U\2W\2Y\2[\2\3\2\25\4\2%\'BB\3\2"+
		"\63;\3\2\62;\4\2DDdd\4\2QQqq\4\2ZZzz\4\2RRrr\4\2--//\3\2bb\4\2$$^^\4\2"+
		"\f\f\17\17\4\2\13\13\"\"\5\2\13\f\17\17\"\"\5\2\f\f\17\17))\13\2$$))^"+
		"^cdhhppttvvxx\3\2\629\5\2\62;CHch\3\2\62\63\4\2GGgg\59\2\62\2;\2\u0662"+
		"\2\u066b\2\u06f2\2\u06fb\2\u07c2\2\u07cb\2\u0968\2\u0971\2\u09e8\2\u09f1"+
		"\2\u0a68\2\u0a71\2\u0ae8\2\u0af1\2\u0b68\2\u0b71\2\u0be8\2\u0bf1\2\u0c68"+
		"\2\u0c71\2\u0ce8\2\u0cf1\2\u0d68\2\u0d71\2\u0de8\2\u0df1\2\u0e52\2\u0e5b"+
		"\2\u0ed2\2\u0edb\2\u0f22\2\u0f2b\2\u1042\2\u104b\2\u1092\2\u109b\2\u17e2"+
		"\2\u17eb\2\u1812\2\u181b\2\u1948\2\u1951\2\u19d2\2\u19db\2\u1a82\2\u1a8b"+
		"\2\u1a92\2\u1a9b\2\u1b52\2\u1b5b\2\u1bb2\2\u1bbb\2\u1c42\2\u1c4b\2\u1c52"+
		"\2\u1c5b\2\ua622\2\ua62b\2\ua8d2\2\ua8db\2\ua902\2\ua90b\2\ua9d2\2\ua9db"+
		"\2\ua9f2\2\ua9fb\2\uaa52\2\uaa5b\2\uabf2\2\uabfb\2\uff12\2\uff1b\2\u04a2"+
		"\3\u04ab\3\u1068\3\u1071\3\u10f2\3\u10fb\3\u1138\3\u1141\3\u11d2\3\u11db"+
		"\3\u12f2\3\u12fb\3\u1452\3\u145b\3\u14d2\3\u14db\3\u1652\3\u165b\3\u16c2"+
		"\3\u16cb\3\u1732\3\u173b\3\u18e2\3\u18eb\3\u1c52\3\u1c5b\3\u1d52\3\u1d5b"+
		"\3\u6a62\3\u6a6b\3\u6b52\3\u6b5b\3\ud7d0\3\ud801\3\ue952\3\ue95b\3\u024b"+
		"\2C\2\\\2c\2|\2\u00ac\2\u00ac\2\u00b7\2\u00b7\2\u00bc\2\u00bc\2\u00c2"+
		"\2\u00d8\2\u00da\2\u00f8\2\u00fa\2\u02c3\2\u02c8\2\u02d3\2\u02e2\2\u02e6"+
		"\2\u02ee\2\u02ee\2\u02f0\2\u02f0\2\u0372\2\u0376\2\u0378\2\u0379\2\u037c"+
		"\2\u037f\2\u0381\2\u0381\2\u0388\2\u0388\2\u038a\2\u038c\2\u038e\2\u038e"+
		"\2\u0390\2\u03a3\2\u03a5\2\u03f7\2\u03f9\2\u0483\2\u048c\2\u0531\2\u0533"+
		"\2\u0558\2\u055b\2\u055b\2\u0563\2\u0589\2\u05d2\2\u05ec\2\u05f2\2\u05f4"+
		"\2\u0622\2\u064c\2\u0670\2\u0671\2\u0673\2\u06d5\2\u06d7\2\u06d7\2\u06e7"+
		"\2\u06e8\2\u06f0\2\u06f1\2\u06fc\2\u06fe\2\u0701\2\u0701\2\u0712\2\u0712"+
		"\2\u0714\2\u0731\2\u074f\2\u07a7\2\u07b3\2\u07b3\2\u07cc\2\u07ec\2\u07f6"+
		"\2\u07f7\2\u07fc\2\u07fc\2\u0802\2\u0817\2\u081c\2\u081c\2\u0826\2\u0826"+
		"\2\u082a\2\u082a\2\u0842\2\u085a\2\u0862\2\u086c\2\u08a2\2\u08b6\2\u08b8"+
		"\2\u08bf\2\u0906\2\u093b\2\u093f\2\u093f\2\u0952\2\u0952\2\u095a\2\u0963"+
		"\2\u0973\2\u0982\2\u0987\2\u098e\2\u0991\2\u0992\2\u0995\2\u09aa\2\u09ac"+
		"\2\u09b2\2\u09b4\2\u09b4\2\u09b8\2\u09bb\2\u09bf\2\u09bf\2\u09d0\2\u09d0"+
		"\2\u09de\2\u09df\2\u09e1\2\u09e3\2\u09f2\2\u09f3\2\u09fe\2\u09fe\2\u0a07"+
		"\2\u0a0c\2\u0a11\2\u0a12\2\u0a15\2\u0a2a\2\u0a2c\2\u0a32\2\u0a34\2\u0a35"+
		"\2\u0a37\2\u0a38\2\u0a3a\2\u0a3b\2\u0a5b\2\u0a5e\2\u0a60\2\u0a60\2\u0a74"+
		"\2\u0a76\2\u0a87\2\u0a8f\2\u0a91\2\u0a93\2\u0a95\2\u0aaa\2\u0aac\2\u0ab2"+
		"\2\u0ab4\2\u0ab5\2\u0ab7\2\u0abb\2\u0abf\2\u0abf\2\u0ad2\2\u0ad2\2\u0ae2"+
		"\2\u0ae3\2\u0afb\2\u0afb\2\u0b07\2\u0b0e\2\u0b11\2\u0b12\2\u0b15\2\u0b2a"+
		"\2\u0b2c\2\u0b32\2\u0b34\2\u0b35\2\u0b37\2\u0b3b\2\u0b3f\2\u0b3f\2\u0b5e"+
		"\2\u0b5f\2\u0b61\2\u0b63\2\u0b73\2\u0b73\2\u0b85\2\u0b85\2\u0b87\2\u0b8c"+
		"\2\u0b90\2\u0b92\2\u0b94\2\u0b97\2\u0b9b\2\u0b9c\2\u0b9e\2\u0b9e\2\u0ba0"+
		"\2\u0ba1\2\u0ba5\2\u0ba6\2\u0baa\2\u0bac\2\u0bb0\2\u0bbb\2\u0bd2\2\u0bd2"+
		"\2\u0c07\2\u0c0e\2\u0c10\2\u0c12\2\u0c14\2\u0c2a\2\u0c2c\2\u0c3b\2\u0c3f"+
		"\2\u0c3f\2\u0c5a\2\u0c5c\2\u0c62\2\u0c63\2\u0c82\2\u0c82\2\u0c87\2\u0c8e"+
		"\2\u0c90\2\u0c92\2\u0c94\2\u0caa\2\u0cac\2\u0cb5\2\u0cb7\2\u0cbb\2\u0cbf"+
		"\2\u0cbf\2\u0ce0\2\u0ce0\2\u0ce2\2\u0ce3\2\u0cf3\2\u0cf4\2\u0d07\2\u0d0e"+
		"\2\u0d10\2\u0d12\2\u0d14\2\u0d3c\2\u0d3f\2\u0d3f\2\u0d50\2\u0d50\2\u0d56"+
		"\2\u0d58\2\u0d61\2\u0d63\2\u0d7c\2\u0d81\2\u0d87\2\u0d98\2\u0d9c\2\u0db3"+
		"\2\u0db5\2\u0dbd\2\u0dbf\2\u0dbf\2\u0dc2\2\u0dc8\2\u0e03\2\u0e32\2\u0e34"+
		"\2\u0e35\2\u0e42\2\u0e48\2\u0e83\2\u0e84\2\u0e86\2\u0e86\2\u0e89\2\u0e8a"+
		"\2\u0e8c\2\u0e8c\2\u0e8f\2\u0e8f\2\u0e96\2\u0e99\2\u0e9b\2\u0ea1\2\u0ea3"+
		"\2\u0ea5\2\u0ea7\2\u0ea7\2\u0ea9\2\u0ea9\2\u0eac\2\u0ead\2\u0eaf\2\u0eb2"+
		"\2\u0eb4\2\u0eb5\2\u0ebf\2\u0ebf\2\u0ec2\2\u0ec6\2\u0ec8\2\u0ec8\2\u0ede"+
		"\2\u0ee1\2\u0f02\2\u0f02\2\u0f42\2\u0f49\2\u0f4b\2\u0f6e\2\u0f8a\2\u0f8e"+
		"\2\u1002\2\u102c\2\u1041\2\u1041\2\u1052\2\u1057\2\u105c\2\u105f\2\u1063"+
		"\2\u1063\2\u1067\2\u1068\2\u1070\2\u1072\2\u1077\2\u1083\2\u1090\2\u1090"+
		"\2\u10a2\2\u10c7\2\u10c9\2\u10c9\2\u10cf\2\u10cf\2\u10d2\2\u10fc\2\u10fe"+
		"\2\u124a\2\u124c\2\u124f\2\u1252\2\u1258\2\u125a\2\u125a\2\u125c\2\u125f"+
		"\2\u1262\2\u128a\2\u128c\2\u128f\2\u1292\2\u12b2\2\u12b4\2\u12b7\2\u12ba"+
		"\2\u12c0\2\u12c2\2\u12c2\2\u12c4\2\u12c7\2\u12ca\2\u12d8\2\u12da\2\u1312"+
		"\2\u1314\2\u1317\2\u131a\2\u135c\2\u1382\2\u1391\2\u13a2\2\u13f7\2\u13fa"+
		"\2\u13ff\2\u1403\2\u166e\2\u1671\2\u1681\2\u1683\2\u169c\2\u16a2\2\u16ec"+
		"\2\u16f3\2\u16fa\2\u1702\2\u170e\2\u1710\2\u1713\2\u1722\2\u1733\2\u1742"+
		"\2\u1753\2\u1762\2\u176e\2\u1770\2\u1772\2\u1782\2\u17b5\2\u17d9\2\u17d9"+
		"\2\u17de\2\u17de\2\u1822\2\u1879\2\u1882\2\u1886\2\u1889\2\u18aa\2\u18ac"+
		"\2\u18ac\2\u18b2\2\u18f7\2\u1902\2\u1920\2\u1952\2\u196f\2\u1972\2\u1976"+
		"\2\u1982\2\u19ad\2\u19b2\2\u19cb\2\u1a02\2\u1a18\2\u1a22\2\u1a56\2\u1aa9"+
		"\2\u1aa9\2\u1b07\2\u1b35\2\u1b47\2\u1b4d\2\u1b85\2\u1ba2\2\u1bb0\2\u1bb1"+
		"\2\u1bbc\2\u1be7\2\u1c02\2\u1c25\2\u1c4f\2\u1c51\2\u1c5c\2\u1c7f\2\u1c82"+
		"\2\u1c8a\2\u1ceb\2\u1cee\2\u1cf0\2\u1cf3\2\u1cf7\2\u1cf8\2\u1d02\2\u1dc1"+
		"\2\u1e02\2\u1f17\2\u1f1a\2\u1f1f\2\u1f22\2\u1f47\2\u1f4a\2\u1f4f\2\u1f52"+
		"\2\u1f59\2\u1f5b\2\u1f5b\2\u1f5d\2\u1f5d\2\u1f5f\2\u1f5f\2\u1f61\2\u1f7f"+
		"\2\u1f82\2\u1fb6\2\u1fb8\2\u1fbe\2\u1fc0\2\u1fc0\2\u1fc4\2\u1fc6\2\u1fc8"+
		"\2\u1fce\2\u1fd2\2\u1fd5\2\u1fd8\2\u1fdd\2\u1fe2\2\u1fee\2\u1ff4\2\u1ff6"+
		"\2\u1ff8\2\u1ffe\2\u2073\2\u2073\2\u2081\2\u2081\2\u2092\2\u209e\2\u2104"+
		"\2\u2104\2\u2109\2\u2109\2\u210c\2\u2115\2\u2117\2\u2117\2\u211b\2\u211f"+
		"\2\u2126\2\u2126\2\u2128\2\u2128\2\u212a\2\u212a\2\u212c\2\u212f\2\u2131"+
		"\2\u213b\2\u213e\2\u2141\2\u2147\2\u214b\2\u2150\2\u2150\2\u2185\2\u2186"+
		"\2\u2c02\2\u2c30\2\u2c32\2\u2c60\2\u2c62\2\u2ce6\2\u2ced\2\u2cf0\2\u2cf4"+
		"\2\u2cf5\2\u2d02\2\u2d27\2\u2d29\2\u2d29\2\u2d2f\2\u2d2f\2\u2d32\2\u2d69"+
		"\2\u2d71\2\u2d71\2\u2d82\2\u2d98\2\u2da2\2\u2da8\2\u2daa\2\u2db0\2\u2db2"+
		"\2\u2db8\2\u2dba\2\u2dc0\2\u2dc2\2\u2dc8\2\u2dca\2\u2dd0\2\u2dd2\2\u2dd8"+
		"\2\u2dda\2\u2de0\2\u2e31\2\u2e31\2\u3007\2\u3008\2\u3033\2\u3037\2\u303d"+
		"\2\u303e\2\u3043\2\u3098\2\u309f\2\u30a1\2\u30a3\2\u30fc\2\u30fe\2\u3101"+
		"\2\u3107\2\u3130\2\u3133\2\u3190\2\u31a2\2\u31bc\2\u31f2\2\u3201\2\u3402"+
		"\2\u4db7\2\u4e02\2\u9fec\2\ua002\2\ua48e\2\ua4d2\2\ua4ff\2\ua502\2\ua60e"+
		"\2\ua612\2\ua621\2\ua62c\2\ua62d\2\ua642\2\ua670\2\ua681\2\ua69f\2\ua6a2"+
		"\2\ua6e7\2\ua719\2\ua721\2\ua724\2\ua78a\2\ua78d\2\ua7b0\2\ua7b2\2\ua7b9"+
		"\2\ua7f9\2\ua803\2\ua805\2\ua807\2\ua809\2\ua80c\2\ua80e\2\ua824\2\ua842"+
		"\2\ua875\2\ua884\2\ua8b5\2\ua8f4\2\ua8f9\2\ua8fd\2\ua8fd\2\ua8ff\2\ua8ff"+
		"\2\ua90c\2\ua927\2\ua932\2\ua948\2\ua962\2\ua97e\2\ua986\2\ua9b4\2\ua9d1"+
		"\2\ua9d1\2\ua9e2\2\ua9e6\2\ua9e8\2\ua9f1\2\ua9fc\2\uaa00\2\uaa02\2\uaa2a"+
		"\2\uaa42\2\uaa44\2\uaa46\2\uaa4d\2\uaa62\2\uaa78\2\uaa7c\2\uaa7c\2\uaa80"+
		"\2\uaab1\2\uaab3\2\uaab3\2\uaab7\2\uaab8\2\uaabb\2\uaabf\2\uaac2\2\uaac2"+
		"\2\uaac4\2\uaac4\2\uaadd\2\uaadf\2\uaae2\2\uaaec\2\uaaf4\2\uaaf6\2\uab03"+
		"\2\uab08\2\uab0b\2\uab10\2\uab13\2\uab18\2\uab22\2\uab28\2\uab2a\2\uab30"+
		"\2\uab32\2\uab5c\2\uab5e\2\uab67\2\uab72\2\uabe4\2\uac02\2\ud7a5\2\ud7b2"+
		"\2\ud7c8\2\ud7cd\2\ud7fd\2\uf902\2\ufa6f\2\ufa72\2\ufadb\2\ufb02\2\ufb08"+
		"\2\ufb15\2\ufb19\2\ufb1f\2\ufb1f\2\ufb21\2\ufb2a\2\ufb2c\2\ufb38\2\ufb3a"+
		"\2\ufb3e\2\ufb40\2\ufb40\2\ufb42\2\ufb43\2\ufb45\2\ufb46\2\ufb48\2\ufbb3"+
		"\2\ufbd5\2\ufd3f\2\ufd52\2\ufd91\2\ufd94\2\ufdc9\2\ufdf2\2\ufdfd\2\ufe72"+
		"\2\ufe76\2\ufe78\2\ufefe\2\uff23\2\uff3c\2\uff43\2\uff5c\2\uff68\2\uffc0"+
		"\2\uffc4\2\uffc9\2\uffcc\2\uffd1\2\uffd4\2\uffd9\2\uffdc\2\uffde\2\2\3"+
		"\r\3\17\3(\3*\3<\3>\3?\3A\3O\3R\3_\3\u0082\3\u00fc\3\u0282\3\u029e\3\u02a2"+
		"\3\u02d2\3\u0302\3\u0321\3\u032f\3\u0342\3\u0344\3\u034b\3\u0352\3\u0377"+
		"\3\u0382\3\u039f\3\u03a2\3\u03c5\3\u03ca\3\u03d1\3\u0402\3\u049f\3\u04b2"+
		"\3\u04d5\3\u04da\3\u04fd\3\u0502\3\u0529\3\u0532\3\u0565\3\u0602\3\u0738"+
		"\3\u0742\3\u0757\3\u0762\3\u0769\3\u0802\3\u0807\3\u080a\3\u080a\3\u080c"+
		"\3\u0837\3\u0839\3\u083a\3\u083e\3\u083e\3\u0841\3\u0857\3\u0862\3\u0878"+
		"\3\u0882\3\u08a0\3\u08e2\3\u08f4\3\u08f6\3\u08f7\3\u0902\3\u0917\3\u0922"+
		"\3\u093b\3\u0982\3\u09b9\3\u09c0\3\u09c1\3\u0a02\3\u0a02\3\u0a12\3\u0a15"+
		"\3\u0a17\3\u0a19\3\u0a1b\3\u0a35\3\u0a62\3\u0a7e\3\u0a82\3\u0a9e\3\u0ac2"+
		"\3\u0ac9\3\u0acb\3\u0ae6\3\u0b02\3\u0b37\3\u0b42\3\u0b57\3\u0b62\3\u0b74"+
		"\3\u0b82\3\u0b93\3\u0c02\3\u0c4a\3\u0c82\3\u0cb4\3\u0cc2\3\u0cf4\3\u1005"+
		"\3\u1039\3\u1085\3\u10b1\3\u10d2\3\u10ea\3\u1105\3\u1128\3\u1152\3\u1174"+
		"\3\u1178\3\u1178\3\u1185\3\u11b4\3\u11c3\3\u11c6\3\u11dc\3\u11dc\3\u11de"+
		"\3\u11de\3\u1202\3\u1213\3\u1215\3\u122d\3\u1282\3\u1288\3\u128a\3\u128a"+
		"\3\u128c\3\u128f\3\u1291\3\u129f\3\u12a1\3\u12aa\3\u12b2\3\u12e0\3\u1307"+
		"\3\u130e\3\u1311\3\u1312\3\u1315\3\u132a\3\u132c\3\u1332\3\u1334\3\u1335"+
		"\3\u1337\3\u133b\3\u133f\3\u133f\3\u1352\3\u1352\3\u135f\3\u1363\3\u1402"+
		"\3\u1436\3\u1449\3\u144c\3\u1482\3\u14b1\3\u14c6\3\u14c7\3\u14c9\3\u14c9"+
		"\3\u1582\3\u15b0\3\u15da\3\u15dd\3\u1602\3\u1631\3\u1646\3\u1646\3\u1682"+
		"\3\u16ac\3\u1702\3\u171b\3\u18a2\3\u18e1\3\u1901\3\u1901\3\u1a02\3\u1a02"+
		"\3\u1a0d\3\u1a34\3\u1a3c\3\u1a3c\3\u1a52\3\u1a52\3\u1a5e\3\u1a85\3\u1a88"+
		"\3\u1a8b\3\u1ac2\3\u1afa\3\u1c02\3\u1c0a\3\u1c0c\3\u1c30\3\u1c42\3\u1c42"+
		"\3\u1c74\3\u1c91\3\u1d02\3\u1d08\3\u1d0a\3\u1d0b\3\u1d0d\3\u1d32\3\u1d48"+
		"\3\u1d48\3\u2002\3\u239b\3\u2482\3\u2545\3\u3002\3\u3430\3\u4402\3\u4648"+
		"\3\u6802\3\u6a3a\3\u6a42\3\u6a60\3\u6ad2\3\u6aef\3\u6b02\3\u6b31\3\u6b42"+
		"\3\u6b45\3\u6b65\3\u6b79\3\u6b7f\3\u6b91\3\u6f02\3\u6f46\3\u6f52\3\u6f52"+
		"\3\u6f95\3\u6fa1\3\u6fe2\3\u6fe3\3\u7002\3\u87ee\3\u8802\3\u8af4\3\ub002"+
		"\3\ub120\3\ub172\3\ub2fd\3\ubc02\3\ubc6c\3\ubc72\3\ubc7e\3\ubc82\3\ubc8a"+
		"\3\ubc92\3\ubc9b\3\ud402\3\ud456\3\ud458\3\ud49e\3\ud4a0\3\ud4a1\3\ud4a4"+
		"\3\ud4a4\3\ud4a7\3\ud4a8\3\ud4ab\3\ud4ae\3\ud4b0\3\ud4bb\3\ud4bd\3\ud4bd"+
		"\3\ud4bf\3\ud4c5\3\ud4c7\3\ud507\3\ud509\3\ud50c\3\ud50f\3\ud516\3\ud518"+
		"\3\ud51e\3\ud520\3\ud53b\3\ud53d\3\ud540\3\ud542\3\ud546\3\ud548\3\ud548"+
		"\3\ud54c\3\ud552\3\ud554\3\ud6a7\3\ud6aa\3\ud6c2\3\ud6c4\3\ud6dc\3\ud6de"+
		"\3\ud6fc\3\ud6fe\3\ud716\3\ud718\3\ud736\3\ud738\3\ud750\3\ud752\3\ud770"+
		"\3\ud772\3\ud78a\3\ud78c\3\ud7aa\3\ud7ac\3\ud7c4\3\ud7c6\3\ud7cd\3\ue802"+
		"\3\ue8c6\3\ue902\3\ue945\3\uee02\3\uee05\3\uee07\3\uee21\3\uee23\3\uee24"+
		"\3\uee26\3\uee26\3\uee29\3\uee29\3\uee2b\3\uee34\3\uee36\3\uee39\3\uee3b"+
		"\3\uee3b\3\uee3d\3\uee3d\3\uee44\3\uee44\3\uee49\3\uee49\3\uee4b\3\uee4b"+
		"\3\uee4d\3\uee4d\3\uee4f\3\uee51\3\uee53\3\uee54\3\uee56\3\uee56\3\uee59"+
		"\3\uee59\3\uee5b\3\uee5b\3\uee5d\3\uee5d\3\uee5f\3\uee5f\3\uee61\3\uee61"+
		"\3\uee63\3\uee64\3\uee66\3\uee66\3\uee69\3\uee6c\3\uee6e\3\uee74\3\uee76"+
		"\3\uee79\3\uee7b\3\uee7e\3\uee80\3\uee80\3\uee82\3\uee8b\3\uee8d\3\uee9d"+
		"\3\ueea3\3\ueea5\3\ueea7\3\ueeab\3\ueead\3\ueebd\3\2\4\ua6d8\4\ua702\4"+
		"\ub736\4\ub742\4\ub81f\4\ub822\4\ucea3\4\uceb2\4\uebe2\4\uf802\4\ufa1f"+
		"\4\u013b\2#\2\61\2<\2B\2]\2b\2}\2\u0080\2\u00a3\2\u00ab\2\u00ad\2\u00ae"+
		"\2\u00b0\2\u00b3\2\u00b6\2\u00b6\2\u00b8\2\u00ba\2\u00bd\2\u00bd\2\u00c1"+
		"\2\u00c1\2\u00d9\2\u00d9\2\u00f9\2\u00f9\2\u02c4\2\u02c7\2\u02d4\2\u02e1"+
		"\2\u02e7\2\u02ed\2\u02ef\2\u02ef\2\u02f1\2\u0301\2\u0377\2\u0377\2\u0380"+
		"\2\u0380\2\u0386\2\u0387\2\u0389\2\u0389\2\u03f8\2\u03f8\2\u0484\2\u0484"+
		"\2\u048a\2\u048b\2\u055c\2\u0561\2\u058b\2\u058c\2\u058f\2\u0591\2\u05c0"+
		"\2\u05c0\2\u05c2\2\u05c2\2\u05c5\2\u05c5\2\u05c8\2\u05c8\2\u05f5\2\u05f6"+
		"\2\u0608\2\u0611\2\u061d\2\u061d\2\u0620\2\u0621\2\u066c\2\u066f\2\u06d6"+
		"\2\u06d6\2\u06e0\2\u06e0\2\u06eb\2\u06eb\2\u06ff\2\u0700\2\u0702\2\u070f"+
		"\2\u07f8\2\u07fb\2\u0832\2\u0840\2\u0860\2\u0860\2\u0966\2\u0967\2\u0972"+
		"\2\u0972\2\u09f4\2\u09f5\2\u09fc\2\u09fd\2\u09ff\2\u09ff\2\u0af2\2\u0af3"+
		"\2\u0b72\2\u0b72\2\u0bf5\2\u0bfc\2\u0c81\2\u0c81\2\u0d51\2\u0d51\2\u0d7b"+
		"\2\u0d7b\2\u0df6\2\u0df6\2\u0e41\2\u0e41\2\u0e51\2\u0e51\2\u0e5c\2\u0e5d"+
		"\2\u0f03\2\u0f19\2\u0f1c\2\u0f21\2\u0f36\2\u0f36\2\u0f38\2\u0f38\2\u0f3a"+
		"\2\u0f3a\2\u0f3c\2\u0f3f\2\u0f87\2\u0f87\2\u0fc0\2\u0fc7\2\u0fc9\2\u0fce"+
		"\2\u0fd0\2\u0fdc\2\u104c\2\u1051\2\u10a0\2\u10a1\2\u10fd\2\u10fd\2\u1362"+
		"\2\u136a\2\u1392\2\u139b\2\u1402\2\u1402\2\u166f\2\u1670\2\u169d\2\u169e"+
		"\2\u16ed\2\u16ef\2\u1737\2\u1738\2\u17d6\2\u17d8\2\u17da\2\u17dd\2\u1802"+
		"\2\u180c\2\u1942\2\u1942\2\u1946\2\u1947\2\u19e0\2\u1a01\2\u1a20\2\u1a21"+
		"\2\u1aa2\2\u1aa8\2\u1aaa\2\u1aaf\2\u1ac0\2\u1ac0\2\u1b5c\2\u1b6c\2\u1b76"+
		"\2\u1b7e\2\u1bfe\2\u1c01\2\u1c3d\2\u1c41\2\u1c80\2\u1c81\2\u1cc2\2\u1cc9"+
		"\2\u1cd5\2\u1cd5\2\u1fbf\2\u1fbf\2\u1fc1\2\u1fc3\2\u1fcf\2\u1fd1\2\u1fdf"+
		"\2\u1fe1\2\u1fef\2\u1ff1\2\u1fff\2\u2000\2\u2012\2\u2029\2\u2032\2\u2060"+
		"\2\u207c\2\u2080\2\u208c\2\u2090\2\u20a2\2\u20c1\2\u20df\2\u20e2\2\u20e4"+
		"\2\u20e6\2\u2102\2\u2103\2\u2105\2\u2108\2\u210a\2\u210b\2\u2116\2\u2116"+
		"\2\u2118\2\u211a\2\u2120\2\u2125\2\u2127\2\u2127\2\u2129\2\u2129\2\u212b"+
		"\2\u212b\2\u2130\2\u2130\2\u213c\2\u213d\2\u2142\2\u2146\2\u214c\2\u214f"+
		"\2\u2151\2\u2151\2\u218c\2\u218d\2\u2192\2\u2428\2\u2442\2\u244c\2\u249e"+
		"\2\u24eb\2\u2502\2\u2777\2\u2796\2\u2b75\2\u2b78\2\u2b97\2\u2b9a\2\u2bbb"+
		"\2\u2bbf\2\u2bca\2\u2bcc\2\u2bd4\2\u2bee\2\u2bf1\2\u2ce7\2\u2cec\2\u2cfb"+
		"\2\u2cfe\2\u2d00\2\u2d01\2\u2d72\2\u2d72\2\u2e02\2\u2e30\2\u2e32\2\u2e4b"+
		"\2\u2e82\2\u2e9b\2\u2e9d\2\u2ef5\2\u2f02\2\u2fd7\2\u2ff2\2\u2ffd\2\u3003"+
		"\2\u3006\2\u300a\2\u3022\2\u3032\2\u3032\2\u3038\2\u3039\2\u303f\2\u3041"+
		"\2\u309d\2\u309e\2\u30a2\2\u30a2\2\u30fd\2\u30fd\2\u3192\2\u3193\2\u3198"+
		"\2\u31a1\2\u31c2\2\u31e5\2\u3202\2\u3220\2\u322c\2\u3249\2\u3252\2\u3252"+
		"\2\u3262\2\u3281\2\u328c\2\u32b2\2\u32c2\2\u3300\2\u3302\2\u3401\2\u4dc2"+
		"\2\u4e01\2\ua492\2\ua4c8\2\ua500\2\ua501\2\ua60f\2\ua611\2\ua672\2\ua675"+
		"\2\ua680\2\ua680\2\ua6f4\2\ua6f9\2\ua702\2\ua718\2\ua722\2\ua723\2\ua78b"+
		"\2\ua78c\2\ua82a\2\ua82d\2\ua838\2\ua83b\2\ua876\2\ua879\2\ua8d0\2\ua8d1"+
		"\2\ua8fa\2\ua8fc\2\ua8fe\2\ua8fe\2\ua930\2\ua931\2\ua961\2\ua961\2\ua9c3"+
		"\2\ua9cf\2\ua9e0\2\ua9e1\2\uaa5e\2\uaa61\2\uaa79\2\uaa7b\2\uaae0\2\uaae1"+
		"\2\uaaf2\2\uaaf3\2\uab5d\2\uab5d\2\uabed\2\uabed\2\ufb2b\2\ufb2b\2\ufbb4"+
		"\2\ufbc3\2\ufd40\2\ufd41\2\ufdfe\2\ufdff\2\ufe12\2\ufe1b\2\ufe32\2\ufe54"+
		"\2\ufe56\2\ufe68\2\ufe6a\2\ufe6d\2\uff03\2\uff11\2\uff1c\2\uff22\2\uff3d"+
		"\2\uff42\2\uff5d\2\uff67\2\uffe2\2\uffe8\2\uffea\2\ufff0\2\ufffe\2\uffff"+
		"\2\u0102\3\u0104\3\u0139\3\u0141\3\u017b\3\u018b\3\u018e\3\u0190\3\u0192"+
		"\3\u019d\3\u01a2\3\u01a2\3\u01d2\3\u01fe\3\u03a1\3\u03a1\3\u03d2\3\u03d2"+
		"\3\u0571\3\u0571\3\u0859\3\u0859\3\u0879\3\u087a\3\u0921\3\u0921\3\u0941"+
		"\3\u0941\3\u0a52\3\u0a5a\3\u0a81\3\u0a81\3\u0aca\3\u0aca\3\u0af2\3\u0af8"+
		"\3\u0b3b\3\u0b41\3\u0b9b\3\u0b9e\3\u1049\3\u104f\3\u10bd\3\u10be\3\u10c0"+
		"\3\u10c3\3\u1142\3\u1145\3\u1176\3\u1177\3\u11c7\3\u11cb\3\u11cf\3\u11cf"+
		"\3\u11dd\3\u11dd\3\u11df\3\u11e1\3\u123a\3\u123f\3\u12ab\3\u12ab\3\u144d"+
		"\3\u1451\3\u145d\3\u145d\3\u145f\3\u145f\3\u14c8\3\u14c8\3\u15c3\3\u15d9"+
		"\3\u1643\3\u1645\3\u1662\3\u166e\3\u173e\3\u1741\3\u1a41\3\u1a48\3\u1a9c"+
		"\3\u1a9e\3\u1aa0\3\u1aa4\3\u1c43\3\u1c47\3\u1c72\3\u1c73\3\u2472\3\u2476"+
		"\3\u6a70\3\u6a71\3\u6af7\3\u6af7\3\u6b39\3\u6b41\3\u6b46\3\u6b47\3\ubc9e"+
		"\3\ubc9e\3\ubca1\3\ubca1\3\ud002\3\ud0f7\3\ud102\3\ud128\3\ud12b\3\ud166"+
		"\3\ud16c\3\ud16e\3\ud185\3\ud186\3\ud18e\3\ud1ab\3\ud1b0\3\ud1ea\3\ud202"+
		"\3\ud243\3\ud247\3\ud247\3\ud302\3\ud358\3\ud6c3\3\ud6c3\3\ud6dd\3\ud6dd"+
		"\3\ud6fd\3\ud6fd\3\ud717\3\ud717\3\ud737\3\ud737\3\ud751\3\ud751\3\ud771"+
		"\3\ud771\3\ud78b\3\ud78b\3\ud7ab\3\ud7ab\3\ud7c5\3\ud7c5\3\ud802\3\uda01"+
		"\3\uda39\3\uda3c\3\uda6f\3\uda76\3\uda78\3\uda85\3\uda87\3\uda8d\3\ue960"+
		"\3\ue961\3\ueef2\3\ueef3\3\uf002\3\uf02d\3\uf032\3\uf095\3\uf0a2\3\uf0b0"+
		"\3\uf0b3\3\uf0c1\3\uf0c3\3\uf0d1\3\uf0d3\3\uf0f7\3\uf112\3\uf130\3\uf132"+
		"\3\uf16d\3\uf172\3\uf1ae\3\uf1e8\3\uf204\3\uf212\3\uf23d\3\uf242\3\uf24a"+
		"\3\uf252\3\uf253\3\uf262\3\uf267\3\uf302\3\uf6d6\3\uf6e2\3\uf6ee\3\uf6f2"+
		"\3\uf6fa\3\uf702\3\uf775\3\uf782\3\uf7d6\3\uf802\3\uf80d\3\uf812\3\uf849"+
		"\3\uf852\3\uf85b\3\uf862\3\uf889\3\uf892\3\uf8af\3\uf902\3\uf90d\3\uf912"+
		"\3\uf940\3\uf942\3\uf94e\3\uf952\3\uf96d\3\uf982\3\uf999\3\uf9c2\3\uf9c2"+
		"\3\uf9d2\3\uf9e8\3\u01d5\2\3\3\2\2\2\2\5\3\2\2\2\2\7\3\2\2\2\2\t\3\2\2"+
		"\2\2\13\3\2\2\2\2\r\3\2\2\2\2\17\3\2\2\2\2\21\3\2\2\2\2\23\3\2\2\2\2\25"+
		"\3\2\2\2\2\27\3\2\2\2\2\31\3\2\2\2\2\33\3\2\2\2\2\35\3\2\2\2\2\37\3\2"+
		"\2\2\2!\3\2\2\2\2#\3\2\2\2\2)\3\2\2\2\2-\3\2\2\2\2/\3\2\2\2\2\61\3\2\2"+
		"\2\2\63\3\2\2\2\2\65\3\2\2\2\2\67\3\2\2\2\29\3\2\2\2\2;\3\2\2\2\2=\3\2"+
		"\2\2\2?\3\2\2\2\2A\3\2\2\2\2C\3\2\2\2\2E\3\2\2\2\3]\3\2\2\2\5_\3\2\2\2"+
		"\7a\3\2\2\2\tc\3\2\2\2\13e\3\2\2\2\rg\3\2\2\2\17i\3\2\2\2\21r\3\2\2\2"+
		"\23u\3\2\2\2\25\u0080\3\2\2\2\27\u008f\3\2\2\2\31\u0091\3\2\2\2\33\u009b"+
		"\3\2\2\2\35\u00a7\3\2\2\2\37\u00b3\3\2\2\2!\u00c5\3\2\2\2#\u00c7\3\2\2"+
		"\2%\u00eb\3\2\2\2\'\u00ed\3\2\2\2)\u00f8\3\2\2\2+\u00fc\3\2\2\2-\u0103"+
		"\3\2\2\2/\u0107\3\2\2\2\61\u0109\3\2\2\2\63\u010e\3\2\2\2\65\u0113\3\2"+
		"\2\2\67\u011a\3\2\2\29\u0125\3\2\2\2;\u012e\3\2\2\2=\u0138\3\2\2\2?\u0146"+
		"\3\2\2\2A\u0152\3\2\2\2C\u0159\3\2\2\2E\u0160\3\2\2\2G\u016a\3\2\2\2I"+
		"\u016c\3\2\2\2K\u0188\3\2\2\2M\u0192\3\2\2\2O\u0194\3\2\2\2Q\u0196\3\2"+
		"\2\2S\u0198\3\2\2\2U\u01a0\3\2\2\2W\u01a2\3\2\2\2Y\u01a4\3\2\2\2[\u01a6"+
		"\3\2\2\2]^\7*\2\2^\4\3\2\2\2_`\7+\2\2`\6\3\2\2\2ab\7}\2\2b\b\3\2\2\2c"+
		"d\7\177\2\2d\n\3\2\2\2ef\7]\2\2f\f\3\2\2\2gh\7_\2\2h\16\3\2\2\2ij\7.\2"+
		"\2j\20\3\2\2\2ks\5\3\2\2ls\5\5\3\2ms\5\13\6\2ns\5\r\7\2os\5\7\4\2ps\5"+
		"\t\5\2qs\5\17\b\2rk\3\2\2\2rl\3\2\2\2rm\3\2\2\2rn\3\2\2\2ro\3\2\2\2rp"+
		"\3\2\2\2rq\3\2\2\2s\22\3\2\2\2tv\t\2\2\2ut\3\2\2\2uv\3\2\2\2vw\3\2\2\2"+
		"w|\5U+\2x{\5U+\2y{\5W,\2zx\3\2\2\2zy\3\2\2\2{~\3\2\2\2|z\3\2\2\2|}\3\2"+
		"\2\2}\24\3\2\2\2~|\3\2\2\2\177\u0081\5[.\2\u0080\177\3\2\2\2\u0081\u0082"+
		"\3\2\2\2\u0082\u0080\3\2\2\2\u0082\u0083\3\2\2\2\u0083\26\3\2\2\2\u0084"+
		"\u0090\7\62\2\2\u0085\u008c\t\3\2\2\u0086\u0088\7a\2\2\u0087\u0086\3\2"+
		"\2\2\u0087\u0088\3\2\2\2\u0088\u0089\3\2\2\2\u0089\u008b\t\4\2\2\u008a"+
		"\u0087\3\2\2\2\u008b\u008e\3\2\2\2\u008c\u008a\3\2\2\2\u008c\u008d\3\2"+
		"\2\2\u008d\u0090\3\2\2\2\u008e\u008c\3\2\2\2\u008f\u0084\3\2\2\2\u008f"+
		"\u0085\3\2\2\2\u0090\30\3\2\2\2\u0091\u0092\7\62\2\2\u0092\u0097\t\5\2"+
		"\2\u0093\u0095\7a\2\2\u0094\u0093\3\2\2\2\u0094\u0095\3\2\2\2\u0095\u0096"+
		"\3\2\2\2\u0096\u0098\5Q)\2\u0097\u0094\3\2\2\2\u0098\u0099\3\2\2\2\u0099"+
		"\u0097\3\2\2\2\u0099\u009a\3\2\2\2\u009a\32\3\2\2\2\u009b\u009d\7\62\2"+
		"\2\u009c\u009e\t\6\2\2\u009d\u009c\3\2\2\2\u009d\u009e\3\2\2\2\u009e\u00a3"+
		"\3\2\2\2\u009f\u00a1\7a\2\2\u00a0\u009f\3\2\2\2\u00a0\u00a1\3\2\2\2\u00a1"+
		"\u00a2\3\2\2\2\u00a2\u00a4\5M\'\2\u00a3\u00a0\3\2\2\2\u00a4\u00a5\3\2"+
		"\2\2\u00a5\u00a3\3\2\2\2\u00a5\u00a6\3\2\2\2\u00a6\34\3\2\2\2\u00a7\u00a8"+
		"\7\62\2\2\u00a8\u00ad\t\7\2\2\u00a9\u00ab\7a\2\2\u00aa\u00a9\3\2\2\2\u00aa"+
		"\u00ab\3\2\2\2\u00ab\u00ac\3\2\2\2\u00ac\u00ae\5O(\2\u00ad\u00aa\3\2\2"+
		"\2\u00ae\u00af\3\2\2\2\u00af\u00ad\3\2\2\2\u00af\u00b0\3\2\2\2\u00b0\36"+
		"\3\2\2\2\u00b1\u00b4\5!\21\2\u00b2\u00b4\5#\22\2\u00b3\u00b1\3\2\2\2\u00b3"+
		"\u00b2\3\2\2\2\u00b4 \3\2\2\2\u00b5\u00be\5K&\2\u00b6\u00b8\7\60\2\2\u00b7"+
		"\u00b9\5K&\2\u00b8\u00b7\3\2\2\2\u00b8\u00b9\3\2\2\2\u00b9\u00bb\3\2\2"+
		"\2\u00ba\u00bc\5S*\2\u00bb\u00ba\3\2\2\2\u00bb\u00bc\3\2\2\2\u00bc\u00bf"+
		"\3\2\2\2\u00bd\u00bf\5S*\2\u00be\u00b6\3\2\2\2\u00be\u00bd\3\2\2\2\u00bf"+
		"\u00c6\3\2\2\2\u00c0\u00c1\7\60\2\2\u00c1\u00c3\5K&\2\u00c2\u00c4\5S*"+
		"\2\u00c3\u00c2\3\2\2\2\u00c3\u00c4\3\2\2\2\u00c4\u00c6\3\2\2\2\u00c5\u00b5"+
		"\3\2\2\2\u00c5\u00c0\3\2\2\2\u00c6\"\3\2\2\2\u00c7\u00c8\7\62\2\2\u00c8"+
		"\u00c9\t\7\2\2\u00c9\u00ca\5%\23\2\u00ca\u00cb\5\'\24\2\u00cb$\3\2\2\2"+
		"\u00cc\u00ce\7a\2\2\u00cd\u00cc\3\2\2\2\u00cd\u00ce\3\2\2\2\u00ce\u00cf"+
		"\3\2\2\2\u00cf\u00d1\5O(\2\u00d0\u00cd\3\2\2\2\u00d1\u00d2\3\2\2\2\u00d2"+
		"\u00d0\3\2\2\2\u00d2\u00d3\3\2\2\2\u00d3\u00de\3\2\2\2\u00d4\u00db\7\60"+
		"\2\2\u00d5\u00d7\7a\2\2\u00d6\u00d5\3\2\2\2\u00d6\u00d7\3\2\2\2\u00d7"+
		"\u00d8\3\2\2\2\u00d8\u00da\5O(\2\u00d9\u00d6\3\2\2\2\u00da\u00dd\3\2\2"+
		"\2\u00db\u00d9\3\2\2\2\u00db\u00dc\3\2\2\2\u00dc\u00df\3\2\2\2\u00dd\u00db"+
		"\3\2\2\2\u00de\u00d4\3\2\2\2\u00de\u00df\3\2\2\2\u00df\u00ec\3\2\2\2\u00e0"+
		"\u00e1\7\60\2\2\u00e1\u00e8\5O(\2\u00e2\u00e4\7a\2\2\u00e3\u00e2\3\2\2"+
		"\2\u00e3\u00e4\3\2\2\2\u00e4\u00e5\3\2\2\2\u00e5\u00e7\5O(\2\u00e6\u00e3"+
		"\3\2\2\2\u00e7\u00ea\3\2\2\2\u00e8\u00e6\3\2\2\2\u00e8\u00e9\3\2\2\2\u00e9"+
		"\u00ec\3\2\2\2\u00ea\u00e8\3\2\2\2\u00eb\u00d0\3\2\2\2\u00eb\u00e0\3\2"+
		"\2\2\u00ec&\3\2\2\2\u00ed\u00ef\t\b\2\2\u00ee\u00f0\t\t\2\2\u00ef\u00ee"+
		"\3\2\2\2\u00ef\u00f0\3\2\2\2\u00f0\u00f1\3\2\2\2\u00f1\u00f2\5K&\2\u00f2"+
		"(\3\2\2\2\u00f3\u00f9\5\27\f\2\u00f4\u00f9\5\31\r\2\u00f5\u00f9\5\33\16"+
		"\2\u00f6\u00f9\5\35\17\2\u00f7\u00f9\5\37\20\2\u00f8\u00f3\3\2\2\2\u00f8"+
		"\u00f4\3\2\2\2\u00f8\u00f5\3\2\2\2\u00f8\u00f6\3\2\2\2\u00f8\u00f7\3\2"+
		"\2\2\u00f9\u00fa\3\2\2\2\u00fa\u00fb\7k\2\2\u00fb*\3\2\2\2\u00fc\u00ff"+
		"\7)\2\2\u00fd\u0100\5G$\2\u00fe\u0100\5/\30\2\u00ff\u00fd\3\2\2\2\u00ff"+
		"\u00fe\3\2\2\2\u0100\u0101\3\2\2\2\u0101\u0102\7)\2\2\u0102,\3\2\2\2\u0103"+
		"\u0104\5+\26\2\u0104.\3\2\2\2\u0105\u0108\5\61\31\2\u0106\u0108\5\63\32"+
		"\2\u0107\u0105\3\2\2\2\u0107\u0106\3\2\2\2\u0108\60\3\2\2\2\u0109\u010a"+
		"\7^\2\2\u010a\u010b\5M\'\2\u010b\u010c\5M\'\2\u010c\u010d\5M\'\2\u010d"+
		"\62\3\2\2\2\u010e\u010f\7^\2\2\u010f\u0110\7z\2\2\u0110\u0111\5O(\2\u0111"+
		"\u0112\5O(\2\u0112\64\3\2\2\2\u0113\u0114\7^\2\2\u0114\u0115\7w\2\2\u0115"+
		"\u0116\5O(\2\u0116\u0117\5O(\2\u0117\u0118\5O(\2\u0118\u0119\5O(\2\u0119"+
		"\66\3\2\2\2\u011a\u011b\7^\2\2\u011b\u011c\7W\2\2\u011c\u011d\5O(\2\u011d"+
		"\u011e\5O(\2\u011e\u011f\5O(\2\u011f\u0120\5O(\2\u0120\u0121\5O(\2\u0121"+
		"\u0122\5O(\2\u0122\u0123\5O(\2\u0123\u0124\5O(\2\u01248\3\2\2\2\u0125"+
		"\u0129\7b\2\2\u0126\u0128\n\n\2\2\u0127\u0126\3\2\2\2\u0128\u012b\3\2"+
		"\2\2\u0129\u0127\3\2\2\2\u0129\u012a\3\2\2\2\u012a\u012c\3\2\2\2\u012b"+
		"\u0129\3\2\2\2\u012c\u012d\7b\2\2\u012d:\3\2\2\2\u012e\u0133\7$\2\2\u012f"+
		"\u0132\n\13\2\2\u0130\u0132\5I%\2\u0131\u012f\3\2\2\2\u0131\u0130\3\2"+
		"\2\2\u0132\u0135\3\2\2\2\u0133\u0131\3\2\2\2\u0133\u0134\3\2\2\2\u0134"+
		"\u0136\3\2\2\2\u0135\u0133\3\2\2\2\u0136\u0137\7$\2\2\u0137<\3\2\2\2\u0138"+
		"\u0139\7\61\2\2\u0139\u013a\7,\2\2\u013a\u013e\3\2\2\2\u013b\u013d\13"+
		"\2\2\2\u013c\u013b\3\2\2\2\u013d\u0140\3\2\2\2\u013e\u013f\3\2\2\2\u013e"+
		"\u013c\3\2\2\2\u013f\u0141\3\2\2\2\u0140\u013e\3\2\2\2\u0141\u0142\7,"+
		"\2\2\u0142\u0143\7\61\2\2\u0143\u0144\3\2\2\2\u0144\u0145\b\37\2\2\u0145"+
		">\3\2\2\2\u0146\u0147\7\61\2\2\u0147\u0148\7\61\2\2\u0148\u014c\3\2\2"+
		"\2\u0149\u014b\n\f\2\2\u014a\u0149\3\2\2\2\u014b\u014e\3\2\2\2\u014c\u014a"+
		"\3\2\2\2\u014c\u014d\3\2\2\2\u014d\u014f\3\2\2\2\u014e\u014c\3\2\2\2\u014f"+
		"\u0150\b \2\2\u0150@\3\2\2\2\u0151\u0153\t\r\2\2\u0152\u0151\3\2\2\2\u0153"+
		"\u0154\3\2\2\2\u0154\u0152\3\2\2\2\u0154\u0155\3\2\2\2\u0155\u0156\3\2"+
		"\2\2\u0156\u0157\b!\2\2\u0157B\3\2\2\2\u0158\u015a\t\f\2\2\u0159\u0158"+
		"\3\2\2\2\u015a\u015b\3\2\2\2\u015b\u0159\3\2\2\2\u015b\u015c\3\2\2\2\u015c"+
		"\u015d\3\2\2\2\u015d\u015e\b\"\2\2\u015eD\3\2\2\2\u015f\u0161\t\16\2\2"+
		"\u0160\u015f\3\2\2\2\u0161\u0162\3\2\2\2\u0162\u0160\3\2\2\2\u0162\u0163"+
		"\3\2\2\2\u0163\u0164\3\2\2\2\u0164\u0165\b#\2\2\u0165F\3\2\2\2\u0166\u016b"+
		"\n\17\2\2\u0167\u016b\5\65\33\2\u0168\u016b\5\67\34\2\u0169\u016b\5I%"+
		"\2\u016a\u0166\3\2\2\2\u016a\u0167\3\2\2\2\u016a\u0168\3\2\2\2\u016a\u0169"+
		"\3\2\2\2\u016bH\3\2\2\2\u016c\u0186\7^\2\2\u016d\u016e\7w\2\2\u016e\u016f"+
		"\5O(\2\u016f\u0170\5O(\2\u0170\u0171\5O(\2\u0171\u0172\5O(\2\u0172\u0187"+
		"\3\2\2\2\u0173\u0174\7W\2\2\u0174\u0175\5O(\2\u0175\u0176\5O(\2\u0176"+
		"\u0177\5O(\2\u0177\u0178\5O(\2\u0178\u0179\5O(\2\u0179\u017a\5O(\2\u017a"+
		"\u017b\5O(\2\u017b\u017c\5O(\2\u017c\u0187\3\2\2\2\u017d\u0187\t\20\2"+
		"\2\u017e\u017f\5M\'\2\u017f\u0180\5M\'\2\u0180\u0181\5M\'\2\u0181\u0187"+
		"\3\2\2\2\u0182\u0183\7z\2\2\u0183\u0184\5O(\2\u0184\u0185\5O(\2\u0185"+
		"\u0187\3\2\2\2\u0186\u016d\3\2\2\2\u0186\u0173\3\2\2\2\u0186\u017d\3\2"+
		"\2\2\u0186\u017e\3\2\2\2\u0186\u0182\3\2\2\2\u0187J\3\2\2\2\u0188\u018f"+
		"\t\4\2\2\u0189\u018b\7a\2\2\u018a\u0189\3\2\2\2\u018a\u018b\3\2\2\2\u018b"+
		"\u018c\3\2\2\2\u018c\u018e\t\4\2\2\u018d\u018a\3\2\2\2\u018e\u0191\3\2"+
		"\2\2\u018f\u018d\3\2\2\2\u018f\u0190\3\2\2\2\u0190L\3\2\2\2\u0191\u018f"+
		"\3\2\2\2\u0192\u0193\t\21\2\2\u0193N\3\2\2\2\u0194\u0195\t\22\2\2\u0195"+
		"P\3\2\2\2\u0196\u0197\t\23\2\2\u0197R\3\2\2\2\u0198\u019a\t\24\2\2\u0199"+
		"\u019b\t\t\2\2\u019a\u0199\3\2\2\2\u019a\u019b\3\2\2\2\u019b\u019c\3\2"+
		"\2\2\u019c\u019d\5K&\2\u019dT\3\2\2\2\u019e\u01a1\5Y-\2\u019f\u01a1\7"+
		"a\2\2\u01a0\u019e\3\2\2\2\u01a0\u019f\3\2\2\2\u01a1V\3\2\2\2\u01a2\u01a3"+
		"\t\25\2\2\u01a3X\3\2\2\2\u01a4\u01a5\t\26\2\2\u01a5Z\3\2\2\2\u01a6\u01a7"+
		"\t\27\2\2\u01a7\\\3\2\2\2\62\2ruz|\u0082\u0087\u008c\u008f\u0094\u0099"+
		"\u009d\u00a0\u00a5\u00aa\u00af\u00b3\u00b8\u00bb\u00be\u00c3\u00c5\u00cd"+
		"\u00d2\u00d6\u00db\u00de\u00e3\u00e8\u00eb\u00ef\u00f8\u00ff\u0107\u0129"+
		"\u0131\u0133\u013e\u014c\u0154\u015b\u0162\u016a\u0186\u018a\u018f\u019a"+
		"\u01a0\3\2\3\2";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}