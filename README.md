
# Signal LLM Bot

A Go bot that integrates with Signal (via the signal-cli REST API) and Google Gemini to provide automated messaging.

## Quick Start

1. **Start the Signal REST API**

   You must run the [signal-cli-rest-api](https://github.com/bbernhard/signal-cli-rest-api) Docker container before starting this bot. Follow the instructions in their README to get it running.

2. **Configure Environment**

   Copy `.env.sample` to `.env` and fill in your credentials and settings. Do not commit your real secrets.

3. **Run the Bot**

   ```sh
   go run ./cmd/main.go
   ```

---

**Note:** This bot will not work unless the Signal REST API is running and reachable.

## Usage

Once the bot is running, it will listen for incoming messages on the configured Signal number. You can interact with the bot by sending messages, and it will respond based on the logic defined in the bot's implementation.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.
