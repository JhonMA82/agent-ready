use ratatui::{backend::CrosstermBackend, Terminal};

fn main() -> std::io::Result<()> {
    let stdout = std::io::stdout();
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;
    terminal.draw(|_| {})?;
    Ok(())
}
