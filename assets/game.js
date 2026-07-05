class MemoryKana {
	constructor(kana, csrfToken) {
		this.csrfToken = csrfToken;

		// grid and content
		this.grid = document.querySelector(".mk-grid");
		this.tiles;

		// previously clicked tile
		this.clicked = null;

		// dialog
		this.dialog = document.querySelector('#mk-dialog');
		this.playerScore = document.querySelector('#player-score');

		// timer
		this.timer = document.querySelector(".mk-timer");
		this.timerStarted;
		this.timerHandle;

		// score
		this.score = 0;
		this.maxScore = 12;

		this.initGame(kana);
	}

	initGame(kana) {
		this.createTiles();

		switch (kana) {
		case "hiragana": this.kana = this.hiragana; break;
		case "katakana": this.kana = this.katakana; break;
		default: throw new Error("invalid kana");
		}

		this.populateTiles(this.kana);
		this.timerStarted = false;
		this.tiles.forEach(tile => (new Tile(tile)).enableClick(this.handleTileClick.bind(this)));
	}

	startTimer() {
		let seconds = 0;
		let minutes = 0;

		this.timerStarted = true;

		let format = (time) => {
			return (time < 10) ? "0" + time : time;
		};

		// display timer
		this.timerHandle = setInterval(() => {
			seconds++;

			if (seconds >= 60) {
				seconds = 0;
				minutes++;
			}

			this.timer.innerHTML = format(minutes) + ":" + format(seconds);
		}, 1000);
	}

	handleTileClick(tile) {
		// init timer on first click
		if (!this.timerStarted) {
			this.startTimer();
		}

		// show kana/romaji
		tile.addClass("clicked")

		if (!this.clicked) {
			this.clicked = tile
			return;
		}

		const kanaPair = tile.type == "kana" && this.kana[this.clicked.pair] == tile.pair;
		const romajiPair = tile.type == "romaji" && this.kana[tile.pair] == this.clicked.pair;

		// pair found!
		if (kanaPair || romajiPair) {
			// permanently show
			this.clicked.addClass("show");
			this.clicked.disableClick();

			tile.addClass("show");
			tile.disableClick();

			// increase score
			this.score++;
		} else {
			// hide clicked items after 200ms
			const prevTile = this.clicked;
			setTimeout(() => {
				prevTile.removeClass("clicked");
				tile.removeClass("clicked");
			}, 200);
		}

		// reset clicked
		this.clicked = null;

		// game over with victory
		if (this.score == this.maxScore) {
			clearInterval(this.timerHandle);

			const elapsedTime = this.timer.innerHTML;

			// create form dinamically and submit
			// reason: make redirect from Go work automatically
			{
				const form = document.createElement("form");
				form.style.display = "none";
				form.method = "POST";
				form.action = "/scoreboard";

				const playerTimeInput = document.createElement("input");
				playerTimeInput.name = "player-time";
				playerTimeInput.value = elapsedTime;
				form.appendChild(playerTimeInput);
				document.body.appendChild(form);

				const csrfInput = document.createElement("input");
				csrfInput.name = "gorilla.csrf.Token";
				csrfInput.value = this.csrfToken;
				form.appendChild(csrfInput);
				document.body.appendChild(form);

				form.submit();
			}
		}
	}

	createTiles() {
		let numOfTiles = 24;
		for (let i = 0; i < numOfTiles; i++) {
			let li = document.createElement("li");
			let span = document.createElement("span");
			this.grid.appendChild(li).appendChild(span);
		}
		this.tiles = Array.prototype.slice.call(this.grid.querySelectorAll("li"))
	}

	populateTiles(kanaType) {
		let temp = this.tiles.slice();
		while (temp.length > 0) {
			// random remove from array
			let kana = temp.splice(this.randomNumber(0, temp.length), 1)[0].children[0];
			let romaji = temp.splice(this.randomNumber(0, temp.length), 1)[0].children[0];
			// loop if duplicate is found
			let prop = this.randomProperty(kanaType);
			while (this.checkDuplicate(prop)) {
				prop = this.randomProperty(kanaType);
			}

			kana.setAttribute("data-pair", kanaType[prop]);
			kana.setAttribute("data-type", "kana");
			kana.innerHTML = prop;

			romaji.setAttribute("data-pair", prop);
			romaji.setAttribute("data-type", "romaji");
			romaji.innerHTML = kanaType[prop];
		}
	}

	checkDuplicate(test) {
		for (let i = 0, len = this.tiles.length; i < len; i++) {
			let span = this.tiles[i].children[0];
			if (span.innerHTML == test) return true;
		}
		return false;
	}

	randomProperty(obj) {
		let keys = Object.keys(obj);
		return keys[keys.length * Math.random() << 0];
	}

	randomNumber(min, max) {
		return Math.floor(Math.random() * (max - min)) + min;
	}

	hiragana = {
		"あ": "a", "い": "i", "う": "u", "え": "e", "お": "o",
		"か": "ka", "き": "ki", "く": "ku", "け": "ke", "こ": "ko",
		"さ": "sa", "し": "shi", "す": "su", "せ": "se", "そ": "so",
		"た": "ta", "ち": "chi", "つ": "tsu", "て": "te", "と": "to",
		"な": "na", "に": "ni", "ぬ": "nu", "ね": "ne", "の": "no",
		"は": "ha", "ひ": "hi", "ふ": "fu", "へ": "he", "ほ": "ho",
		"ま": "ma", "み": "mi", "む": "mu", "め": "me", "も": "mo",
		"や": "ya", "ゆ": "yu", "よ": "yo",
		"ら": "ra", "り": "ri", "る": "ru", "れ": "re", "ろ": "ro",
		"わ": "wa", "を": "wo",
		"ん": "n"
	};

	katakana = {
		"ア": "a", "イ": "i", "ウ": "u", "エ": "e", "オ": "o",
		"カ": "ka", "キ": "ki", "ク": "ku", "ケ": "ke", "コ": "ko",
		"サ": "sa", "シ": "shi", "ス": "su", "セ": "se", "ソ": "so",
		"タ": "ta", "チ": "chi", "ツ": "tsu", "テ": "te", "ト": "to",
		"ナ": "na", "ニ": "ni", "ヌ": "nu", "ネ": "ne", "ノ": "no",
		"ハ": "ha", "ヒ": "hi", "フ": "fu", "ヘ": "he", "ホ": "ho",
		"マ": "ma", "ミ": "mi", "ム": "mu", "メ": "me", "モ": "mo",
		"ヤ": "ya", "ユ": "yu", "ヨ": "yo",
		"ラ": "ra", "リ": "ri", "ル": "ru", "レ": "re", "ロ": "ro",
		"ワ": "wa", "ヲ": "wo",
		"ン": "n"
	};
}

class Tile {
	constructor(element) {
		this.element = element;
		this.pair = this.element.children[0].dataset.pair;
		this.type = this.element.children[0].dataset.type;
	}

	enableClick(handler) {
		this.element.addEventListener('click', () => handler(this));
	}

	disableClick() {
		this.element.removeEventListener('click', () => handler(this));
	}

	addClass(className) {
		this.element.children[0].classList.add(className);
	}

	removeClass(className) {
		this.element.children[0].classList.remove(className);
	}
}
