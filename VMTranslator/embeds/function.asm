(%[1]s)
@%[2]d
D=A
@%[1]s$INIT_END
D;JLE
@R13
M=D
(%[1]s$INIT_LOOP)
@SP
A=M
M=0
@SP
M=M+1
@R13
MD=M-1
@%[1]s$INIT_LOOP
D;JGT
(%[1]s$INIT_END)
