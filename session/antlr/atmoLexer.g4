lexer grammar atmoLexer;

/* taken from https://github.com/antlr/grammars-v4/blob/master/golang/GoLexer.g4 and then tweaked:
 * removed keywords, operators, most of punctuation; modified rules: IDENTIFIER; added rules:
 * OPERATOR
 */

// $antlr-format alignTrailingComments true, columnLimit 180, maxEmptyLinesToKeep 1, reflowComments false, useTab false
// $antlr-format allowShortRulesOnASingleLine true, allowShortBlocksOnASingleLine true, minEmptyLines 0, alignSemicolons ownLine
// $antlr-format alignColons trailing, singleLineOverrulesHangingColon true, alignLexerCommands true, alignLabels true, alignTrailers true

IDENTIFIER : ( ('@' | '$' | '%' | '#')? LETTER (LETTER | UNICODE_DIGIT)*);

OPERATOR : UNICODE_OPISH+;

// Punctuation

L_PAREN   : '(';
R_PAREN   : ')';
L_CURLY   : '{';
R_CURLY   : '}';
L_BRACKET : '[';
R_BRACKET : ']';
COMMA     : ',';

// Number literals

DECIMAL_LIT : ('0' | [1-9] ('_'? [0-9])*);
BINARY_LIT  : '0' [bB] ('_'? BIN_DIGIT)+;
OCTAL_LIT   : '0' [oO]? ('_'? OCTAL_DIGIT)+;
HEX_LIT     : '0' [xX] ('_'? HEX_DIGIT)+;

FLOAT_LIT : (DECIMAL_FLOAT_LIT | HEX_FLOAT_LIT);

DECIMAL_FLOAT_LIT : DECIMALS ('.' DECIMALS? EXPONENT? | EXPONENT) | '.' DECIMALS EXPONENT?;

HEX_FLOAT_LIT : '0' [xX] HEX_MANTISSA HEX_EXPONENT;

fragment HEX_MANTISSA : ('_'? HEX_DIGIT)+ ('.' ( '_'? HEX_DIGIT)*)? | '.' HEX_DIGIT ('_'? HEX_DIGIT)*;

fragment HEX_EXPONENT : [pP] [+-]? DECIMALS;

IMAGINARY_LIT : (DECIMAL_LIT | BINARY_LIT | OCTAL_LIT | HEX_LIT | FLOAT_LIT) 'i';

// Rune literals

fragment RUNE : '\'' (UNICODE_VALUE | BYTE_VALUE) '\''; //: '\'' (~[\n\\] | ESCAPED_VALUE) '\'';

RUNE_LIT : RUNE;

BYTE_VALUE : OCTAL_BYTE_VALUE | HEX_BYTE_VALUE;

OCTAL_BYTE_VALUE : '\\' OCTAL_DIGIT OCTAL_DIGIT OCTAL_DIGIT;

HEX_BYTE_VALUE : '\\' 'x' HEX_DIGIT HEX_DIGIT;

LITTLE_U_VALUE : '\\' 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;

BIG_U_VALUE : '\\' 'U' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;

// String literals

RAW_STRING_LIT         : '`' ~'`'* '`';
INTERPRETED_STRING_LIT : '"' (~["\\] | ESCAPED_VALUE)* '"';

// Hidden tokens

WS           : [ \t]+        -> channel(HIDDEN);
COMMENT      : '/*' .*? '*/' -> channel(HIDDEN);
LINE_COMMENT : '//' ~[\r\n]* -> channel(HIDDEN);
TERMINATOR   : [\r\n]+       -> channel(HIDDEN);

fragment UNICODE_VALUE : ~[\r\n'] | LITTLE_U_VALUE | BIG_U_VALUE | ESCAPED_VALUE;

// Fragments

fragment ESCAPED_VALUE:
    '\\' (
        'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT
        | 'U' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT
        | [abfnrtv\\'"]
        | OCTAL_DIGIT OCTAL_DIGIT OCTAL_DIGIT
        | 'x' HEX_DIGIT HEX_DIGIT
    )
;

fragment DECIMALS : [0-9] ('_'? [0-9])*;

fragment OCTAL_DIGIT : [0-7];

fragment HEX_DIGIT : [0-9a-fA-F];

fragment BIN_DIGIT : [01];

fragment EXPONENT : [eE] [+-]? DECIMALS;

fragment LETTER : UNICODE_LETTER | '_';

//[\p{Nd}] matches a digit zero through nine in any script except ideographic scripts
fragment UNICODE_DIGIT : [\p{Nd}];
//[\p{L}] matches any kind of letter from any language
fragment UNICODE_LETTER : [\p{L}];

fragment UNICODE_OPISH : [\p{Sm}\p{Sc}\p{Sk}];
