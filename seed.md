
### espeficicacoes.ferramentas-agnos

### especificacoes.ferramentas-agnos-descrição:
Cli ja instalado no pc, que serve para construir aplicacoes.
### especificacoes.ferramentas-agnos-modo-de-uso:
use o modo backoffice com **backoffice-init** 

### especificacoes.idiona:
o sistema deve ser feito todo em ingles, desde o codigo, ate a interface de uso, porem a nivel interface (api tambem totalmente em ingles), deve ser permitido colocar o idioma em portugues.

### especificacoes.api
deve haver apis para controlar a aplicacao programaticamente.


### objetivo:
construir um erp financeiro completo.


### funcionalidades

### funcionalidades.cadastro-de-contas
permitir cadastrar contas corrente, contas de investimento, carteiras,etc.. 
### funcionalidades.cadastro-de-contas.requisito 
duas empresas podem ter despesas e receitas da mesma conta, o sistema deve ter 
mecanismos de visualizar a conta absoluta (valor de caixa) e o valor por empresa.


### funcionalidades.cadastro-de-terceiros.
terceiros podem ser clientes, fornecedores, ou ambos.

### funcionalidades.cadastro-de-empresas
cadastro de empresas, tudo deve girar em torno de empresas, inclusive se o usuario for usar para uso pessoal, ele deve ser uma empresa,ele vai cadastrar como (pessoal) por exemplo (a criterio do cliente), o sistema nao deve diferenciar pessoas de empresas, tudo no sistema deve ser empresas

### funcionalidades.cadastro-de-contas-a-pagar-e-contas-a-receber
deve ser possivel cadastrar transacoes, e marcar se estao pagas/recebidas, a data do lancamento e data da efetivavao 

### funcionalidades.cadastro-de-budgets.
deve se permitir cadastrar budges, para o mes, ou por periodo, e isso deve refletir nos paineis de controle.

### funcionalidades.relatorios
devem possuir relatorios gerenciais, como fluxo de caixa, contas a pagar e receber, relatorio por empresa, relatorio por conta, etc, assim como mecanismos para filtrar essas coisas.

### funcionalidades.saldo-projetado.
o usuario deve pdoer ver os saldos projetados, com e sem os budges, o saldo projetado de daqui um ano, e todas as despesas e receitas que terao ate la, + os budges


### funcionalidades.cartoes-de-credito
permitir cartoes de credito no sistema.

