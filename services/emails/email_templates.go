package emails

import "os"

func GetNewRecruiterOrCompanyUserEmail(validationToken string, email string) string {
	return `
	Confirmação de email			
	
	Olá,
	
	Um novo usuário com email ` + email + ` se cadastrou como recruiter/ empresa.
	Para confirmar o email, clique ou copie e cole no seu navegador favorito o link abaixo:
	`+os.Getenv("BASE_UI_HOST")+`/confirmacao?token=` + validationToken + `

	Atenciosamente,
	Equipe Vagas para Jr.
	contato@vagasprajr.com.br
	`
}

func GetCompanyRecruiterAskingLinksEmail() string {
	return `
	Olá,
	
	Seja bem vinda(o) à @vagasprajr.
	Estamos muito felizes em ter você conosco e só precisamos de mais um passo para validar seu cadastro como recruiter/ empresa.

	Responda esse email e no corpo do email, por favor informe o nome da empresa que você representa com o link de suas redes sociais como recruiter e o link do seu perfil no LinkedIn.
	Esta validação é necessária para garantir que apenas empresas e recrutadores possam cadastrar vagas no: ` + os.Getenv("BASE_UI_HOST") + `

	Estamos ansiosos para ter você conosco postando vagas para os nossos candidatos.

	Atenciosamente,
	Equipe Vagas para Jr.
	contato@vagasprajr.com.br
	`
}

func GetWelcomeEmail(validationToken string) string {
	return `
	Confirmação de email			
	
	Olá,
	
	Seja bem vinda(o) ao Vagas para Jr. 
	Para confirmar seu email, clique ou copie e cole no seu navegador favorito o link abaixo:
	`+os.Getenv("BASE_UI_HOST")+`/confirmacao?token=` + validationToken + `

	Atenciosamente,
	Equipe Vagas para Jr.
	contato@vagasprajr.com.br	`
}

func ReceiptSend(userEmail string, receipt string) string {
	return `
	Envio de recibo
	
	O usuário ` + userEmail + ` enviou um recibo para o Clube @vagasprajr.

	Recibo: ` + receipt + `

	Atenciosamente,
	Equipe Vagas para Jr.
	contato@vagasprajr.com.br	`
}