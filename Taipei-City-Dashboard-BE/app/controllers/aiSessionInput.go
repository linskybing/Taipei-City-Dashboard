package controllers

type AIChatSessionCreateInput struct {
	Title string `json:"title" binding:"omitempty,max=120"`
}

type AIChatSessionRenameInput struct {
	Title string `json:"title" binding:"required,max=120"`
}
