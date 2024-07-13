// Generated from /home/_/c/go/_/atmo/session/antlr/Atmo.g4 by ANTLR 4.9.2
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast"})
public class AtmoParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.9.2", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		IDENTIFIER=1, OPERATOR=2, L_PAREN=3, R_PAREN=4, L_CURLY=5, R_CURLY=6, 
		L_BRACKET=7, R_BRACKET=8, COMMA=9, DECIMAL_LIT=10, BINARY_LIT=11, OCTAL_LIT=12, 
		HEX_LIT=13, FLOAT_LIT=14, DECIMAL_FLOAT_LIT=15, HEX_FLOAT_LIT=16, IMAGINARY_LIT=17, 
		RUNE_LIT=18, BYTE_VALUE=19, OCTAL_BYTE_VALUE=20, HEX_BYTE_VALUE=21, LITTLE_U_VALUE=22, 
		BIG_U_VALUE=23, RAW_STRING_LIT=24, INTERPRETED_STRING_LIT=25, COMMENT=26, 
		LINE_COMMENT=27, WS=28, TERMINATOR=29, WS_NL=30;
	public static final int
		RULE_comment = 0, RULE_expr = 1, RULE_ident = 2, RULE_lit = 3;
	private static String[] makeRuleNames() {
		return new String[] {
			"comment", "expr", "ident", "lit"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, null, null, "'('", "')'", "'{'", "'}'", "'['", "']'", "','"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "IDENTIFIER", "OPERATOR", "L_PAREN", "R_PAREN", "L_CURLY", "R_CURLY", 
			"L_BRACKET", "R_BRACKET", "COMMA", "DECIMAL_LIT", "BINARY_LIT", "OCTAL_LIT", 
			"HEX_LIT", "FLOAT_LIT", "DECIMAL_FLOAT_LIT", "HEX_FLOAT_LIT", "IMAGINARY_LIT", 
			"RUNE_LIT", "BYTE_VALUE", "OCTAL_BYTE_VALUE", "HEX_BYTE_VALUE", "LITTLE_U_VALUE", 
			"BIG_U_VALUE", "RAW_STRING_LIT", "INTERPRETED_STRING_LIT", "COMMENT", 
			"LINE_COMMENT", "WS", "TERMINATOR", "WS_NL"
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

	@Override
	public String getGrammarFileName() { return "Atmo.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public AtmoParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	public static class CommentContext extends ParserRuleContext {
		public TerminalNode LINE_COMMENT() { return getToken(AtmoParser.LINE_COMMENT, 0); }
		public TerminalNode COMMENT() { return getToken(AtmoParser.COMMENT, 0); }
		public CommentContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_comment; }
	}

	public final CommentContext comment() throws RecognitionException {
		CommentContext _localctx = new CommentContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_comment);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(8);
			_la = _input.LA(1);
			if ( !(_la==COMMENT || _la==LINE_COMMENT) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public static class ExprContext extends ParserRuleContext {
		public ExprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expr; }
	 
		public ExprContext() { }
		public void copyFrom(ExprContext ctx) {
			super.copyFrom(ctx);
		}
	}
	public static class IdentExprContext extends ExprContext {
		public IdentContext ident() {
			return getRuleContext(IdentContext.class,0);
		}
		public IdentExprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	public static class LitExprContext extends ExprContext {
		public LitContext lit() {
			return getRuleContext(LitContext.class,0);
		}
		public LitExprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	public static class CallFormExprContext extends ExprContext {
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public CallFormExprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	public static class ParensExprContext extends ExprContext {
		public TerminalNode L_PAREN() { return getToken(AtmoParser.L_PAREN, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode R_PAREN() { return getToken(AtmoParser.R_PAREN, 0); }
		public ParensExprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	public static class SquareBracketExprContext extends ExprContext {
		public TerminalNode L_BRACKET() { return getToken(AtmoParser.L_BRACKET, 0); }
		public TerminalNode R_BRACKET() { return getToken(AtmoParser.R_BRACKET, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(AtmoParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(AtmoParser.COMMA, i);
		}
		public SquareBracketExprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	public static class CurlyBracesExprContext extends ExprContext {
		public TerminalNode L_CURLY() { return getToken(AtmoParser.L_CURLY, 0); }
		public TerminalNode R_CURLY() { return getToken(AtmoParser.R_CURLY, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(AtmoParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(AtmoParser.COMMA, i);
		}
		public CurlyBracesExprContext(ExprContext ctx) { copyFrom(ctx); }
	}

	public final ExprContext expr() throws RecognitionException {
		return expr(0);
	}

	private ExprContext expr(int _p) throws RecognitionException {
		ParserRuleContext _parentctx = _ctx;
		int _parentState = getState();
		ExprContext _localctx = new ExprContext(_ctx, _parentState);
		ExprContext _prevctx = _localctx;
		int _startState = 2;
		enterRecursionRule(_localctx, 2, RULE_expr, _p);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(37);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case L_PAREN:
				{
				_localctx = new ParensExprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;

				setState(11);
				match(L_PAREN);
				setState(12);
				expr(0);
				setState(13);
				match(R_PAREN);
				}
				break;
			case L_BRACKET:
				{
				_localctx = new SquareBracketExprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(15);
				match(L_BRACKET);
				setState(21);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while ((((_la) & ~0x3f) == 0 && ((1L << _la) & ((1L << IDENTIFIER) | (1L << OPERATOR) | (1L << L_PAREN) | (1L << L_CURLY) | (1L << L_BRACKET) | (1L << DECIMAL_LIT) | (1L << BINARY_LIT) | (1L << OCTAL_LIT) | (1L << HEX_LIT) | (1L << FLOAT_LIT) | (1L << IMAGINARY_LIT) | (1L << RUNE_LIT) | (1L << RAW_STRING_LIT) | (1L << INTERPRETED_STRING_LIT))) != 0)) {
					{
					{
					setState(16);
					expr(0);
					setState(17);
					match(COMMA);
					}
					}
					setState(23);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(24);
				match(R_BRACKET);
				}
				break;
			case L_CURLY:
				{
				_localctx = new CurlyBracesExprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(25);
				match(L_CURLY);
				setState(31);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while ((((_la) & ~0x3f) == 0 && ((1L << _la) & ((1L << IDENTIFIER) | (1L << OPERATOR) | (1L << L_PAREN) | (1L << L_CURLY) | (1L << L_BRACKET) | (1L << DECIMAL_LIT) | (1L << BINARY_LIT) | (1L << OCTAL_LIT) | (1L << HEX_LIT) | (1L << FLOAT_LIT) | (1L << IMAGINARY_LIT) | (1L << RUNE_LIT) | (1L << RAW_STRING_LIT) | (1L << INTERPRETED_STRING_LIT))) != 0)) {
					{
					{
					setState(26);
					expr(0);
					setState(27);
					match(COMMA);
					}
					}
					setState(33);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(34);
				match(R_CURLY);
				}
				break;
			case IDENTIFIER:
			case OPERATOR:
				{
				_localctx = new IdentExprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(35);
				ident();
				}
				break;
			case DECIMAL_LIT:
			case BINARY_LIT:
			case OCTAL_LIT:
			case HEX_LIT:
			case FLOAT_LIT:
			case IMAGINARY_LIT:
			case RUNE_LIT:
			case RAW_STRING_LIT:
			case INTERPRETED_STRING_LIT:
				{
				_localctx = new LitExprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(36);
				lit();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			_ctx.stop = _input.LT(-1);
			setState(47);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,4,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					if ( _parseListeners!=null ) triggerExitRuleEvent();
					_prevctx = _localctx;
					{
					{
					_localctx = new CallFormExprContext(new ExprContext(_parentctx, _parentState));
					pushNewRecursionContext(_localctx, _startState, RULE_expr);
					setState(39);
					if (!(precpred(_ctx, 3))) throw new FailedPredicateException(this, "precpred(_ctx, 3)");
					setState(41); 
					_errHandler.sync(this);
					_alt = 1;
					do {
						switch (_alt) {
						case 1:
							{
							{
							setState(40);
							expr(0);
							}
							}
							break;
						default:
							throw new NoViableAltException(this);
						}
						setState(43); 
						_errHandler.sync(this);
						_alt = getInterpreter().adaptivePredict(_input,3,_ctx);
					} while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
					}
					} 
				}
				setState(49);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,4,_ctx);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			unrollRecursionContexts(_parentctx);
		}
		return _localctx;
	}

	public static class IdentContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(AtmoParser.IDENTIFIER, 0); }
		public TerminalNode OPERATOR() { return getToken(AtmoParser.OPERATOR, 0); }
		public IdentContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ident; }
	}

	public final IdentContext ident() throws RecognitionException {
		IdentContext _localctx = new IdentContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_ident);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(50);
			_la = _input.LA(1);
			if ( !(_la==IDENTIFIER || _la==OPERATOR) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public static class LitContext extends ParserRuleContext {
		public TerminalNode RUNE_LIT() { return getToken(AtmoParser.RUNE_LIT, 0); }
		public TerminalNode RAW_STRING_LIT() { return getToken(AtmoParser.RAW_STRING_LIT, 0); }
		public TerminalNode INTERPRETED_STRING_LIT() { return getToken(AtmoParser.INTERPRETED_STRING_LIT, 0); }
		public TerminalNode IMAGINARY_LIT() { return getToken(AtmoParser.IMAGINARY_LIT, 0); }
		public TerminalNode FLOAT_LIT() { return getToken(AtmoParser.FLOAT_LIT, 0); }
		public TerminalNode DECIMAL_LIT() { return getToken(AtmoParser.DECIMAL_LIT, 0); }
		public TerminalNode BINARY_LIT() { return getToken(AtmoParser.BINARY_LIT, 0); }
		public TerminalNode OCTAL_LIT() { return getToken(AtmoParser.OCTAL_LIT, 0); }
		public TerminalNode HEX_LIT() { return getToken(AtmoParser.HEX_LIT, 0); }
		public LitContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_lit; }
	}

	public final LitContext lit() throws RecognitionException {
		LitContext _localctx = new LitContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_lit);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(52);
			_la = _input.LA(1);
			if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & ((1L << DECIMAL_LIT) | (1L << BINARY_LIT) | (1L << OCTAL_LIT) | (1L << HEX_LIT) | (1L << FLOAT_LIT) | (1L << IMAGINARY_LIT) | (1L << RUNE_LIT) | (1L << RAW_STRING_LIT) | (1L << INTERPRETED_STRING_LIT))) != 0)) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public boolean sempred(RuleContext _localctx, int ruleIndex, int predIndex) {
		switch (ruleIndex) {
		case 1:
			return expr_sempred((ExprContext)_localctx, predIndex);
		}
		return true;
	}
	private boolean expr_sempred(ExprContext _localctx, int predIndex) {
		switch (predIndex) {
		case 0:
			return precpred(_ctx, 3);
		}
		return true;
	}

	public static final String _serializedATN =
		"\3\u608b\ua72a\u8133\ub9ed\u417c\u3be7\u7786\u5964\3 9\4\2\t\2\4\3\t\3"+
		"\4\4\t\4\4\5\t\5\3\2\3\2\3\3\3\3\3\3\3\3\3\3\3\3\3\3\3\3\3\3\7\3\26\n"+
		"\3\f\3\16\3\31\13\3\3\3\3\3\3\3\3\3\3\3\7\3 \n\3\f\3\16\3#\13\3\3\3\3"+
		"\3\3\3\5\3(\n\3\3\3\3\3\6\3,\n\3\r\3\16\3-\7\3\60\n\3\f\3\16\3\63\13\3"+
		"\3\4\3\4\3\5\3\5\3\5\2\3\4\6\2\4\6\b\2\5\3\2\34\35\3\2\3\4\5\2\f\20\23"+
		"\24\32\33\2<\2\n\3\2\2\2\4\'\3\2\2\2\6\64\3\2\2\2\b\66\3\2\2\2\n\13\t"+
		"\2\2\2\13\3\3\2\2\2\f\r\b\3\1\2\r\16\7\5\2\2\16\17\5\4\3\2\17\20\7\6\2"+
		"\2\20(\3\2\2\2\21\27\7\t\2\2\22\23\5\4\3\2\23\24\7\13\2\2\24\26\3\2\2"+
		"\2\25\22\3\2\2\2\26\31\3\2\2\2\27\25\3\2\2\2\27\30\3\2\2\2\30\32\3\2\2"+
		"\2\31\27\3\2\2\2\32(\7\n\2\2\33!\7\7\2\2\34\35\5\4\3\2\35\36\7\13\2\2"+
		"\36 \3\2\2\2\37\34\3\2\2\2 #\3\2\2\2!\37\3\2\2\2!\"\3\2\2\2\"$\3\2\2\2"+
		"#!\3\2\2\2$(\7\b\2\2%(\5\6\4\2&(\5\b\5\2\'\f\3\2\2\2\'\21\3\2\2\2\'\33"+
		"\3\2\2\2\'%\3\2\2\2\'&\3\2\2\2(\61\3\2\2\2)+\f\5\2\2*,\5\4\3\2+*\3\2\2"+
		"\2,-\3\2\2\2-+\3\2\2\2-.\3\2\2\2.\60\3\2\2\2/)\3\2\2\2\60\63\3\2\2\2\61"+
		"/\3\2\2\2\61\62\3\2\2\2\62\5\3\2\2\2\63\61\3\2\2\2\64\65\t\3\2\2\65\7"+
		"\3\2\2\2\66\67\t\4\2\2\67\t\3\2\2\2\7\27!\'-\61";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}