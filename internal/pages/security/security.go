package security

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"main/internal/components"
	secengine "main/internal/engine/security"
	"main/internal/i18n"
	sshlib "main/internal/ssh"
	"main/internal/theme"
)

type viewState int

const (
	stateLoading viewState = iota
	stateSummary
	stateFirewall
	statePorts
	stateFail2Ban
	stateAddRule
	stateDeleteRule
	stateBanIP
	stateUnbanIP
)

type Model struct {
	report *secengine.FullAuditReport
	engine *secengine.Engine
	status string
	state  viewState

	// Form inputs
	portInput   textinput.Model
	protoInput  textinput.Model
	actionInput textinput.Model
	idInput     textinput.Model
	ipInput     textinput.Model
	jailInput   textinput.Model

	formStep int
}

func New() Model {
	port := textinput.New()
	port.Placeholder = i18n.T("e.g. 80, 443, 8080")
	port.Focus()

	proto := textinput.New()
	proto.Placeholder = i18n.T("tcp, udp, or any")

	action := textinput.New()
	action.Placeholder = i18n.T("allow or deny")

	idInp := textinput.New()
	idInp.Placeholder = i18n.T("Rule ID")

	ipInp := textinput.New()
	ipInp.Placeholder = i18n.T("IP Address")

	jailInp := textinput.New()
	jailInp.Placeholder = i18n.T("Jail Name (e.g. sshd)")

	return Model{
		status:      i18n.T("Connecting..."),
		state:       stateLoading,
		portInput:   port,
		protoInput:  proto,
		actionInput: action,
		idInput:     idInp,
		ipInput:     ipInp,
		jailInput:   jailInp,
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) IsInputActive() bool {
	return m.state == stateAddRule || m.state == stateDeleteRule || m.state == stateBanIP || m.state == stateUnbanIP
}

type auditReportMsg *secengine.FullAuditReport

func runFullAudit(engine *secengine.Engine) tea.Cmd {
	return func() tea.Msg {
		if engine == nil {
			return nil
		}
		report, err := engine.RunFullAudit()
		if err != nil {
			return nil
		}
		return auditReportMsg(report)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case auditReportMsg:
		m.report = msg
		m.status = i18n.T("Audit complete.")
		m.state = stateSummary
		return m, nil

	case tea.KeyMsg:
		if m.state == stateLoading {
			return m, nil
		}

		if m.state == stateSummary || m.state == stateFirewall || m.state == statePorts || m.state == stateFail2Ban {
			switch msg.String() {
			case "r", "R":
				m.status = i18n.T("Running deep security audit...")
				m.state = stateLoading
				return m, runFullAudit(m.engine)
			case "1":
				m.state = stateSummary
			case "2":
				m.state = stateFirewall
			case "3":
				m.state = statePorts
			case "4":
				m.state = stateFail2Ban
			}
		}

		if m.state == stateFirewall {
			switch msg.String() {
			case "a", "A":
				m.state = stateAddRule
				m.formStep = 0
				m.portInput.SetValue("")
				m.protoInput.SetValue("")
				m.actionInput.SetValue("")
				m.portInput.Focus()
				m.protoInput.Blur()
				m.actionInput.Blur()
			case "d", "D":
				m.state = stateDeleteRule
				m.idInput.SetValue("")
				m.idInput.Focus()
			}
		} else if m.state == stateFail2Ban {
			switch msg.String() {
			case "b", "B":
				m.state = stateBanIP
				m.formStep = 0
				m.ipInput.SetValue("")
				m.jailInput.SetValue("")
				m.ipInput.Focus()
				m.jailInput.Blur()
			case "u", "U":
				m.state = stateUnbanIP
				m.formStep = 0
				m.ipInput.SetValue("")
				m.jailInput.SetValue("")
				m.ipInput.Focus()
				m.jailInput.Blur()
			}
		} else if m.state == stateAddRule {
			switch msg.String() {
			case "esc":
				m.state = stateFirewall
			case "enter":
				if m.formStep == 0 {
					m.formStep++
					m.portInput.Blur()
					m.protoInput.Focus()
				} else if m.formStep == 1 {
					m.formStep++
					m.protoInput.Blur()
					m.actionInput.Focus()
				} else if m.formStep == 2 {
					// submit
					err := m.engine.AddFirewallRule(m.portInput.Value(), m.protoInput.Value(), m.actionInput.Value())
					if err != nil {
						m.status = i18n.T("Error adding rule: ") + err.Error()
					} else {
						m.status = i18n.T("Rule added successfully.")
					}
					m.state = stateLoading
					return m, runFullAudit(m.engine)
				}
			}
			m.portInput, cmd = m.portInput.Update(msg)
			cmds = append(cmds, cmd)
			m.protoInput, cmd = m.protoInput.Update(msg)
			cmds = append(cmds, cmd)
			m.actionInput, cmd = m.actionInput.Update(msg)
			cmds = append(cmds, cmd)

		} else if m.state == stateDeleteRule {
			switch msg.String() {
			case "esc":
				m.state = stateFirewall
			case "enter":
				err := m.engine.RemoveFirewallRule(m.idInput.Value())
				if err != nil {
					m.status = i18n.T("Error deleting rule: ") + err.Error()
				} else {
					m.status = i18n.T("Rule deleted successfully.")
				}
				m.state = stateLoading
				return m, runFullAudit(m.engine)
			}
			m.idInput, cmd = m.idInput.Update(msg)
			cmds = append(cmds, cmd)

		} else if m.state == stateBanIP || m.state == stateUnbanIP {
			switch msg.String() {
			case "esc":
				m.state = stateFail2Ban
			case "enter":
				if m.formStep == 0 {
					m.formStep++
					m.ipInput.Blur()
					m.jailInput.Focus()
				} else if m.formStep == 1 {
					var err error
					if m.state == stateBanIP {
						err = m.engine.BanIP(m.jailInput.Value(), m.ipInput.Value())
					} else {
						err = m.engine.UnbanIP(m.jailInput.Value(), m.ipInput.Value())
					}
					if err != nil {
						m.status = i18n.T("Error: ") + err.Error()
					} else {
						m.status = i18n.T("Success.")
					}
					m.state = stateLoading
					return m, runFullAudit(m.engine)
				}
			}
			m.ipInput, cmd = m.ipInput.Update(msg)
			cmds = append(cmds, cmd)
			m.jailInput, cmd = m.jailInput.Update(msg)
			cmds = append(cmds, cmd)
		}

	case sshlib.ConnectedMsg:
		m.engine = secengine.NewEngine(msg.Client)
		m.status = i18n.T("Running deep security audit...")
		return m, runFullAudit(m.engine)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.state == stateLoading {
		return components.Card(lipgloss.NewStyle().Foreground(theme.Current.Dim).Render(m.status), 60)
	}

	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	label := lipgloss.NewStyle().Foreground(theme.Current.Primary).Width(25).Bold(true)
	text := lipgloss.NewStyle().Foreground(theme.Current.Text)
	dim := lipgloss.NewStyle().Foreground(theme.Current.Dim)

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		components.Title(i18n.T("SECURITY CENTER")),
		"  ",
		dim.Render(i18n.T("[1] Summary   [2] Firewall   [3] Open Ports   [4] Fail2Ban")),
	)

	var content string

	switch m.state {
	case stateSummary:
		var ufwStr string
		if m.report.BasicReport.UFWStatus == "Active" {
			ufwStr = green.Render(i18n.T("Secure (Active)"))
		} else {
			ufwStr = red.Render(i18n.T("VULNERABLE (Inactive/Missing)"))
		}

		var rootStr string
		if m.report.BasicReport.RootLoginEnabled {
			rootStr = red.Render(i18n.T("VULNERABLE (Enabled)"))
		} else {
			rootStr = green.Render(i18n.T("Secure (Disabled)"))
		}

		var passStr string
		if m.report.BasicReport.PasswordAuthEnabled {
			passStr = red.Render(i18n.T("VULNERABLE (Passwords Allowed)"))
		} else {
			passStr = green.Render(i18n.T("Secure (Key Only)"))
		}

		items := []string{
			fmt.Sprintf("%s %s", label.Render(i18n.T("Firewall Engine:")), text.Render(strings.ToUpper(m.report.FirewallType))),
			fmt.Sprintf("%s %s", label.Render(i18n.T("UFW Firewall Status:")), ufwStr),
			fmt.Sprintf("%s %s", label.Render(i18n.T("SSH Root Login:")), rootStr),
			fmt.Sprintf("%s %s", label.Render(i18n.T("SSH Password Auth:")), passStr),
		}

		controls := dim.Render("\n" + i18n.T("Controls: [R] Rerun Audit"))
		content = lipgloss.JoinVertical(lipgloss.Left,
			header, "",
			items[0], items[1], items[2], items[3],
			controls,
		)

	case stateFirewall:
		rulesView := label.Render(i18n.Tf("Active Rules (%s):", m.report.FirewallType)) + "\n"
		if len(m.report.Rules) == 0 {
			rulesView += dim.Render(i18n.T("No rules configured or firewall inactive."))
		} else {
			for _, r := range m.report.Rules {
				rulesView += fmt.Sprintf("[%2s] %-10s %-15s %s\n", r.ID, r.Action, r.To, r.From)
			}
		}

		controls := dim.Render("\n" + i18n.T("Controls: [a] Add Rule  [d] Delete Rule  [r] Rerun Audit"))
		content = lipgloss.JoinVertical(lipgloss.Left,
			header, "",
			rulesView,
			controls,
		)

	case stateAddRule:
		form := lipgloss.JoinVertical(lipgloss.Left,
			label.Render(i18n.T("Add Firewall Rule:")),
			"",
			i18n.T("Port:    ")+m.portInput.View(),
			i18n.T("Proto:   ")+m.protoInput.View(),
			i18n.T("Action:  ")+m.actionInput.View(),
			"",
			dim.Render(i18n.T("[Enter] Next/Submit   [Esc] Cancel")),
		)
		content = lipgloss.JoinVertical(lipgloss.Left, header, "", form)

	case stateDeleteRule:
		form := lipgloss.JoinVertical(lipgloss.Left,
			label.Render(i18n.T("Delete Firewall Rule:")),
			"",
			i18n.T("Rule ID: ")+m.idInput.View(),
			"",
			dim.Render(i18n.T("[Enter] Submit   [Esc] Cancel")),
		)
		content = lipgloss.JoinVertical(lipgloss.Left, header, "", form)

	case statePorts:
		portsView := label.Render(i18n.T("Open Ports & Owning Processes:")) + "\n\n"
		portsView += dim.Render(fmt.Sprintf("%s %s %s", components.PadRight(i18n.T("PROTO"), 8), components.PadRight(i18n.T("ADDRESS:PORT"), 25), components.PadRight(i18n.T("PROCESS"), 20))) + "\n"
		for _, p := range m.report.Ports {
			addrStr := p.Address + ":" + p.Port
			if p.Address == "0.0.0.0" {
				addrStr = red.Render(addrStr)
			} else if p.Address == "*" || p.Address == "::" {
				addrStr = yellow.Render(addrStr)
			}
			portsView += fmt.Sprintf("%-8s %-25s %-20s\n", p.Protocol, addrStr, p.Process)
		}

		controls := dim.Render("\n" + i18n.T("Controls: [r] Rerun Audit"))
		content = lipgloss.JoinVertical(lipgloss.Left, header, "", portsView, controls)

	case stateFail2Ban:
		view := label.Render(i18n.T("Fail2Ban Status:")) + "\n"
		if len(m.report.Jails) == 0 {
			view += dim.Render(i18n.T("Fail2Ban is not active or no jails found.")) + "\n\n"
		} else {
			for _, j := range m.report.Jails {
				banned := strings.Join(j.BannedIPs, ", ")
				if banned == "" {
					banned = "None"
				}
				view += i18n.Tf("- %s: Banned: %s\n", text.Bold(true).Render(j.Name), banned)
			}
			view += "\n"
		}

		view += label.Render(i18n.T("Failed SSH Logins (Last 10 Days):")) + "\n"
		if len(m.report.Logins) == 0 {
			view += dim.Render(i18n.T("No failed logins found.")) + "\n"
		} else {
			view += dim.Render(fmt.Sprintf("%s %s %s", components.PadRight(i18n.T("COUNT"), 6), components.PadRight(i18n.T("IP"), 16), i18n.T("LAST ATTEMPT"))) + "\n"
			for i, l := range m.report.Logins {
				if i >= 8 {
					break // max display 8 to fit screen
				}
				view += fmt.Sprintf("%-6d %-16s %s\n", l.Count, l.IP, l.LastTime)
			}
		}

		controls := dim.Render("\n" + i18n.T("Controls: [b] Ban IP  [u] Unban IP  [r] Rerun Audit"))
		content = lipgloss.JoinVertical(lipgloss.Left, header, "", view, controls)

	case stateBanIP, stateUnbanIP:
		title := i18n.T("Ban IP Address:")
		if m.state == stateUnbanIP {
			title = i18n.T("Unban IP Address:")
		}
		form := lipgloss.JoinVertical(lipgloss.Left,
			label.Render(title),
			"",
			i18n.T("IP Address: ")+m.ipInput.View(),
			i18n.T("Jail Name:  ")+m.jailInput.View(),
			"",
			dim.Render(i18n.T("[Enter] Next/Submit   [Esc] Cancel")),
		)
		content = lipgloss.JoinVertical(lipgloss.Left, header, "", form)
	}

	return components.Card(content, 75)
}

func (m Model) Title() string { return "Security" }
func (m Model) Icon() string  { return "🛡️" }
