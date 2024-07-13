grammar dummy;

// $antlr-format alignTrailingComments true, columnLimit 150, maxEmptyLinesToKeep 1, reflowComments false, useTab false
// $antlr-format allowShortRulesOnASingleLine true, allowShortBlocksOnASingleLine true, minEmptyLines 0, alignSemicolons ownLine
// $antlr-format alignColons trailing, singleLineOverrulesHangingColon true, alignLexerCommands true, alignLabels true, alignTrailers true

operation : NUMBER '+' NUMBER;
foo       : operation | (NUMBER '*' NUMBER);
moofoo    : operation | (NUMBER '/' NUMBER);

NUMBER     : [0-9]+;
WHITESPACE : ' ' -> skip;
