package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type sessionState int

const (
	stateLoading sessionState = iota
	stateSelectSender
	stateSelectReceiver
	stateEnterAmount
	stateSending
	stateSuccess
	stateError
)

type item struct {
	id       uuid.UUID
	name     string
	currency string
}

func (i item) FilterValue() string { return i.name }
func (i item) Title() string       { return i.name }
func (i item) Description() string { return fmt.Sprintf("Currency: %s", i.currency) }

type model struct {
	client       *Client
	state        sessionState
	list         list.Model
	amountInput  textinput.Model
	accounts     []Account
	sender       *Account
	receiver     *Account
	amount       int64
	err          error
	windowWidth  int
	windowHeight int
}

func initialModel(client *Client) model {
	ti := textinput.New()
	ti.Placeholder = "Amount (in cents)"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 20

	return model{
		client:      client,
		state:       stateLoading,
		amountInput: ti,
		list:        list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}
}

func (m model) Init() tea.Cmd {
	return fetchAccounts(m.client)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "esc" && m.state != stateLoading {
			// Backtracking logic could go here
			return m, tea.Quit
		}

	case accountsMsg:
		m.accounts = msg
		items := make([]list.Item, len(m.accounts))
		for i, acc := range m.accounts {
			items[i] = item{id: acc.ID, name: acc.Name, currency: acc.Currency}
		}
		m.list.SetItems(items)
		m.list.Title = "Select Sender"
		m.state = stateSelectSender

	case errMsg:
		m.err = msg
		m.state = stateError
		return m, tea.Quit

	case transactionResultMsg:
		m.state = stateSuccess
		return m, tea.Quit
	}

	switch m.state {
	case stateSelectSender:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					for _, acc := range m.accounts {
						if acc.ID == i.id {
							m.sender = &acc
							break
						}
					}
					m.list.Title = "Select Receiver"
					// Filter out the sender from the list for receiver selection
					var receiverItems []list.Item
					for _, acc := range m.accounts {
						if acc.ID != m.sender.ID {
							receiverItems = append(receiverItems, item{id: acc.ID, name: acc.Name, currency: acc.Currency})
						}
					}
					m.list.SetItems(receiverItems)
					m.list.ResetSelected()
					m.state = stateSelectReceiver
				}
			}
		}
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

	case stateSelectReceiver:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					for _, acc := range m.accounts {
						if acc.ID == i.id {
							m.receiver = &acc
							break
						}
					}
					m.state = stateEnterAmount
				}
			}
		}
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

	case stateEnterAmount:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				amt, err := strconv.ParseInt(m.amountInput.Value(), 10, 64)
				if err == nil && amt > 0 {
					m.amount = amt
					m.state = stateSending
					return m, sendTransaction(m.client, TransactionRequest{
						From:           m.sender.ID,
						To:             m.receiver.ID,
						Currency:       m.sender.Currency, // Assuming sender's currency
						Amount:         m.amount,
						IdempotencyKey: uuid.New(),
					})
				}
			}
		}
		m.amountInput, cmd = m.amountInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return errorMessageStyle(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.state {
	case stateLoading:
		return statusMessageStyle("Loading accounts...")

	case stateSelectSender, stateSelectReceiver:
		return appStyle.Render(m.list.View())

	case stateEnterAmount:
		return appStyle.Render(fmt.Sprintf(
			"Send from %s to %s\n\n%s\n\n%s",
			m.sender.Name,
			m.receiver.Name,
			m.amountInput.View(),
			helpStyle.Render("(Enter amount in cents)"),
		))

	case stateSending:
		return statusMessageStyle("Sending transaction...")

	case stateSuccess:
		return statusMessageStyle(fmt.Sprintf("Successfully sent %d %s from %s to %s!", m.amount, m.sender.Currency, m.sender.Name, m.receiver.Name))
	}

	return ""
}

type accountsMsg []Account
type errMsg error
type transactionResultMsg struct{}

func fetchAccounts(client *Client) tea.Cmd {
	return func() tea.Msg {
		accounts, err := client.GetAccounts()
		if err != nil {
			return errMsg(err)
		}
		return accountsMsg(accounts)
	}
}

func sendTransaction(client *Client, req TransactionRequest) tea.Cmd {
	return func() tea.Msg {
		err := client.SendTransaction(req)
		if err != nil {
			return errMsg(err)
		}
		return transactionResultMsg{}
	}
}
