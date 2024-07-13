// Generated from /home/_/c/go/_/atmo/session/antlr/atmo.g4 by ANTLR 4.9.2
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast"})
public class atmoParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.9.2", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		T__0=1, T__1=2, T__2=3, NUMBER=4, WHITESPACE=5;
	public static final int
		RULE_operation = 0, RULE_foo = 1, RULE_moofoo = 2;
	private static String[] makeRuleNames() {
		return new String[] {
			"operation", "foo", "moofoo"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "'+'", "'*'", "'/'", null, "' '"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, null, null, null, "NUMBER", "WHITESPACE"
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
	public String getGrammarFileName() { return "atmo.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public atmoParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	public static class OperationContext extends ParserRuleContext {
		public List<TerminalNode> NUMBER() { return getTokens(atmoParser.NUMBER); }
		public TerminalNode NUMBER(int i) {
			return getToken(atmoParser.NUMBER, i);
		}
		public OperationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_operation; }
	}

	public final OperationContext operation() throws RecognitionException {
		OperationContext _localctx = new OperationContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_operation);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(6);
			match(NUMBER);
			setState(7);
			match(T__0);
			setState(8);
			match(NUMBER);
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

	public static class FooContext extends ParserRuleContext {
		public OperationContext operation() {
			return getRuleContext(OperationContext.class,0);
		}
		public List<TerminalNode> NUMBER() { return getTokens(atmoParser.NUMBER); }
		public TerminalNode NUMBER(int i) {
			return getToken(atmoParser.NUMBER, i);
		}
		public FooContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_foo; }
	}

	public final FooContext foo() throws RecognitionException {
		FooContext _localctx = new FooContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_foo);
		try {
			setState(14);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,0,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(10);
				operation();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				{
				setState(11);
				match(NUMBER);
				setState(12);
				match(T__1);
				setState(13);
				match(NUMBER);
				}
				}
				break;
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

	public static class MoofooContext extends ParserRuleContext {
		public OperationContext operation() {
			return getRuleContext(OperationContext.class,0);
		}
		public List<TerminalNode> NUMBER() { return getTokens(atmoParser.NUMBER); }
		public TerminalNode NUMBER(int i) {
			return getToken(atmoParser.NUMBER, i);
		}
		public MoofooContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_moofoo; }
	}

	public final MoofooContext moofoo() throws RecognitionException {
		MoofooContext _localctx = new MoofooContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_moofoo);
		try {
			setState(20);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,1,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(16);
				operation();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				{
				setState(17);
				match(NUMBER);
				setState(18);
				match(T__2);
				setState(19);
				match(NUMBER);
				}
				}
				break;
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

	public static final String _serializedATN =
		"\3\u608b\ua72a\u8133\ub9ed\u417c\u3be7\u7786\u5964\3\7\31\4\2\t\2\4\3"+
		"\t\3\4\4\t\4\3\2\3\2\3\2\3\2\3\3\3\3\3\3\3\3\5\3\21\n\3\3\4\3\4\3\4\3"+
		"\4\5\4\27\n\4\3\4\2\2\5\2\4\6\2\2\2\27\2\b\3\2\2\2\4\20\3\2\2\2\6\26\3"+
		"\2\2\2\b\t\7\6\2\2\t\n\7\3\2\2\n\13\7\6\2\2\13\3\3\2\2\2\f\21\5\2\2\2"+
		"\r\16\7\6\2\2\16\17\7\4\2\2\17\21\7\6\2\2\20\f\3\2\2\2\20\r\3\2\2\2\21"+
		"\5\3\2\2\2\22\27\5\2\2\2\23\24\7\6\2\2\24\25\7\5\2\2\25\27\7\6\2\2\26"+
		"\22\3\2\2\2\26\23\3\2\2\2\27\7\3\2\2\2\4\20\26";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}