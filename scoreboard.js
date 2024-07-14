async function displayScoreboard() {
    try {
        const response = await fetch("api/scoreboard");
        const payload = await response.json();

        const scoreboard = document.querySelector('.mk-scoreboard tbody');

        let tableRows = "";
        payload.forEach(playerScore => {
            tableRows += `<tr><td>${playerScore.player}</td><td>${playerScore.score}</td></tr>`
        });
        scoreboard.innerHTML = tableRows;
    } catch (err) {
        console.error(err.message);
    }
}
