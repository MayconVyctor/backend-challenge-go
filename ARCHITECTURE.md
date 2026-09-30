Perguntas gerais do sistema

## Primeira pergunta

imagine que recebemos:

BET
25.00 BRL
wallet = X

Quais coisas precisam acontecer para podermos dizer que essa aposta foi realmente processada?

quais dados precisam existir?

nao focando no usuario mas precisamos de uma carteira existente, no caso entidade com atribustos principalemnte o atributo de carteira 

o que precisa ser validado?

e na carteira tem que ter um saldo >= valor da aposta, talvez se o usuario existe e esta logado com a permissao de fazer a aposta a bet

o que precisa acontecer junto?

precisamos validar o usuario, seu saldo da carteira, seu saldo da carteira em comparacao ao valor da aposta tem que ser maior, verificar se ja nao esta acontecendo uma operacao ao mesmo tempo e se for true apos isso realizar a funcao

o que nao pode acontecer duas vezes?

processar as duas apostas ao mesmo tempo, descontando do usuario duas vezes
