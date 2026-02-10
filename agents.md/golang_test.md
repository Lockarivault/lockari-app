# Testes em Go

## TESTS
Você deve criar uma suíte de testes completa para o projeto Lockari.

**Bibliotecas**
Bibliotecas a serem utilizadas:
- "testing"
- "github.com/stretchr/testify/assert"
- "github.com/stretchr/testify/suite"
- "github.com/stretchr/testify/mock"
- "github.com/stretchr/testify/require"

**Padrões**
- Os testes devem serem escritos com o padrão de testes em Go;
- Devem usar a biblioteca testify;
- Devem ter testsuite;
- Quando contiver interfaces, crie os mocks necessários;
- Quando contiver banco de dados, crie os testes de integração;
- Quando contiver API, crie os testes de integração;
- Quando contiver API, crie os testes de E2E;

**Mocks**
Todos os mocks devem seguir o padrão de mocks em Go;
- diretório tests/mocks/<nome_do_pacote>
- nome do arquivo deve ser <nome_do_pacote>_mock.go
- nome do mock deve ser <NomeDoPacote>Mock

## PERFORMANCE
Devem ser realizados testes de performance para garantir que o sistema esteja performático.

É necessário o desenvolvimento de benchmarks para garantir que o sistema esteja performático.

É importante que os testes de performance sejam realizados com o banco de dados em memória.


## SECURITY
Para os testes de segurança, devem ser utilizados os seguintes padrões:
- Devem ser realizados testes de segurança para garantir que o sistema esteja seguro;
- Devemos usar o **golang.org/x/vuln/cmd/govulncheck@latest** para verificar vulnerabilidades;
- Devemos usar o gitleaks para verificar vulnerabilidades;
- Devemos usar o snyk para verificar vulnerabilidades;


**IMPORTANTE**
É muito importante que os testes sejam realizados, consigam trazer problemas referentes a performance, segurança e testes unitários.

**IMPORTANTE**
Os testes devem ser realizados com o banco de dados em memória.