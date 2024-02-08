# AI-Intelli-Task-Bot -  Slack Bot with AI, Weather Forecast, and Task Management

This Slack bot is a multi-functional assistant designed to enhance communication and organization within Slack channels. It integrates several features including AI-powered question answering, weather forecasting, and task management capabilities.

## Features

### 1. AI Question Answering
- Send any question to Wolfram Alpha and display options based on the response.
- Example usage: `/get answer - Who is the president of India?`

### 2. Weather Forecast
- Get the current weather forecast for a specific city.
- Example usage: `/get weather - New York`

### 3. Task Management
- Create, update, delete tasks with due times.
- View task details including description, status, due time, and creation time.
- Example usage: `/create task - Do homework tomorrow`

## Screenshots

| Details Page | AI Screen | Weather Forecasting| Task Creation |
| ------------ | ------------ | ------------ | ------------ |
![Screenshot 1](https://github.com/VANSHAJ-cpp/AI-Intelli-Task-Bot/assets/93986234/aef05f75-cbc2-44ec-984c-f8e3e1509909) |  ![Screenshot 2](https://github.com/VANSHAJ-cpp/AI-Intelli-Task-Bot/assets/93986234/30a95261-3eba-4633-a2f2-49daca822c98) |  ![Screenshot 3](https://github.com/VANSHAJ-cpp/AI-Intelli-Task-Bot/assets/93986234/452dd082-514d-4175-a6f5-1d8a18756247) | ![Screenshot 4](https://github.com/VANSHAJ-cpp/AI-Intelli-Task-Bot/assets/93986234/e0eb228a-3c99-4735-ad05-0152c465b0f8) |



## Usage

1. **Setup Environment Variables**
   - Copy the `.env.example` file to `.env`.
   - Add your Slack bot token, Slack app token, WIT.ai token, OpenWeather API key, and Wolfram Alpha app ID to the `.env` file.

2. **Installation**
   ```bash
   go mod tidy
   ```

3. **Run the Bot**
   ```bash
   go run main.go
   ```

4. **Commands**
   - `/get answer - <question>`: Ask any question to Wolfram Alpha.
   - `/get weather - <city>`: Get the current weather forecast for a specific city.
   - `/create task - <description> <due_time>`: Create a new task with a due time.
   - `/get task - <task_id>`: Get details of a specific task.
   - `/update task status - <task_id> <status>`: Update the status of a task.
   - `/delete task - <task_id>`: Delete a task.

## Requirements

- Go 1.16 or higher
- Dependencies: `github.com/google/uuid`, `github.com/joho/godotenv`, `github.com/krognol/go-wolfram`, `github.com/shomali11/slacker`, `github.com/slack-go/slack`, `github.com/tidwall/gjson`, `github.com/wit-ai/wit-go/v2`, `golang.org/x/text`

## License

This project is licensed under the [MIT License](LICENSE).
