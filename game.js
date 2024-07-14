class MemoryKana {
    constructor() {
        // grid and content
        this.grid = document.querySelector(".mk-grid");
        this.tiles;

        // modal windows and overlay
        this.modal = document.querySelector(".mk-modal");
        this.overlay = document.querySelector(".mk-overlay");

        // timer
        this.timer = document.querySelector(".mk-timer");
        this.timerStarted;
        this.timerHandle;

        // score
        this.score = 0;
        this.maxScore = 12;

        // initialize game
        this.init(this.hiragana);
    }

    init(kana) {
        this.createTiles();
        this.populateTiles(kana);
        this.timerStarted = false;
        this.initClickEvents();
    }

    startTimer() {
        let min = 0, sec = 1;
        this.timerStarted = true;

        this.timerHandle = setInterval(() => {
            // add leading zeros
            let minString = (min < 10) ? "0"+min : min;
            let secString = (sec < 10) ? "0"+sec : sec;

            // minute increase
            if (sec++ == 60) {
                min++;
                sec = 0;
            }

            // display timer
            this.timer.innerHTML = minString + ":" + secString;
        }, 1000);
    }

    initClickEvents() {
        let clicked = null;
        this.tiles.forEach((item) => {
            item.addEventListener('click', () => {
                // init timer on first click
                if (!this.timerStarted) {
                    this.startTimer();
                }

                // get clicked span
                let span = item.children[0];
                span.classList.add("clicked");

                // clicked 2-times
                if (clicked) {
                    // pair found
                    if (clicked.innerHTML == span.dataset.pair) {
                        // permanently show
                        clicked.classList.add("show");
                        span.classList.add("show");

                        // increase score
                        this.score++;
                    }

                    // game over with victory
                    if (this.score == this.maxScore) {
                        this.modal.classList.add("mk-show");
                        clearInterval(this.timerHandle);
                    }

                    // hide clicked items after 200ms
                    setTimeout(() => {
                        clicked.classList.remove("clicked");
                        span.classList.remove("clicked");
                        clicked = null;
                    }, 200);
                } else {
                    // save clicked item
                    clicked = span;
                }
            });
        });
    };

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
            // add kana and romaji
            kana.setAttribute("data-pair", kanaType[prop]);
            kana.innerHTML = prop;
            romaji.setAttribute("data-pair", prop);
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
        "ハ": "ha", "ヒ": "hi", "フ": "hu", "ヘ": "he", "ホ": "ho",
        "マ": "ma", "ミ": "mi", "ム": "mu", "メ": "me", "モ": "mo",
        "ヤ": "ya", "ユ": "yu", "ヨ": "yo",
        "ラ": "ra", "リ": "ri", "ル": "ru", "レ": "re", "ロ": "ro",
        "ワ": "wa", "ヲ": "wo",
        "ン": "n"
    };
}
