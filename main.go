package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/krognol/go-wolfram"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	wolframClient     *wolfram.Client
	openWeatherAPIKey string
	tasks             map[string]Task
)

type Task struct {
	ID          string
	Description string
	Status      string // e.g., "todo", "in_progress", "completed"
	Assignee    string
	DueTime     time.Time // Due time for the task
	CreatedAt   time.Time
}

func capitalizeString(s string) string {
	return cases.Title(language.English).String(s)
}

func main() {
	godotenv.Load(".env")

	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))

	openWeatherAPIKey = os.Getenv("OPENWEATHER_API_KEY")
	wolframClient = &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	tasks = make(map[string]Task)

	go printCommandEvents(bot.CommandEvents())

	bot.Command("get weather - <city>", &slacker.CommandDefinition{
		Description: "Get the current weather forecast for a specific city",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			city := capitalizeString(request.Param("city"))

			forecast, err := getWeatherForecast(city)
			if err != nil {
				log.Println("Error fetching weather forecast:", err)
				response.Reply("Sorry, I couldn't fetch the weather forecast. Please try again later.")
				return
			}

			message := fmt.Sprintf("Weather forecast for %s:\n", capitalizeString(city))
			message += fmt.Sprintf("Temperature: %.1f°C\n", forecast.Main.Temp)
			if len(forecast.Weather) > 0 {
				message += fmt.Sprintf("Description: %s\n", forecast.Weather[0].Description)
			} else {
				message += "Description: N/A\n" // or any default value you want to use
			}

			response.Reply(message)
		},
	})

	bot.Command("get answer - <message>", &slacker.CommandDefinition{
		Description: "Send any question to Wolfram and display options based on the response",
		Examples:    []string{"who is the president of india"},
		Handler: func(bc slacker.BotContext, r slacker.Request, w slacker.ResponseWriter) {
			query := r.Param("message")

			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})
			data, _ := json.MarshalIndent(msg, "", "    ")
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			answer := value.String()
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)

			if err != nil {
				log.Println("Error querying Wolfram:", err)
				w.Reply("Sorry, I couldn't find an answer to your query.")
				return
			}

			w.Reply(res)

			var actions []slack.AttachmentAction
			if res != "" {
				actions = []slack.AttachmentAction{
					{Name: "satisfied", Text: "Satisfied", Type: "button", Value: "satisfied"},
					{Name: "not_satisfied", Text: "Not Satisfied", Type: "button", Value: "not_satisfied"},
				}
			} else {
				actions = []slack.AttachmentAction{
					{Name: "not_satisfied", Text: "Not Satisfied", Type: "button", Value: "not_satisfied"},
				}
			}

			attachment := slack.Attachment{
				Text:       "Are you satisfied with the answer?",
				Actions:    actions,
				CallbackID: "satisfaction_feedback",
			}

			attachments := []slack.Attachment{attachment}

			replyOptions := []slacker.ReplyOption{
				slacker.WithAttachments(attachments),
			}

			if err := w.Reply("Please provide feedback:", replyOptions...); err != nil {
				log.Println("Error sending message:", err)
			}
		},
	})

	bot.Command("create task - <description> <due_time>", &slacker.CommandDefinition{
		Description: "Create a new task with a due time",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			description := request.Param("description")
			dueTimeString := request.Param("due_time")

			// Parse due time
			dueTime, err := time.Parse("2006-01-02", dueTimeString)
			if err != nil {
				response.Reply("Invalid due time format. Please use YYYY-MM-DD.")
				return
			}

			// Create task
			task := Task{
				ID:          uuid.New().String(), // Generate unique task ID
				Description: description,
				Status:      "todo",
				DueTime:     dueTime,
				CreatedAt:   time.Now(),
			}
			tasks[task.ID] = task

			// Show created task
			showTask(response, task.ID)
		},
	})

	bot.Command("get task - <task_id>", &slacker.CommandDefinition{
		Description: "Get details of a task",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			taskID := request.Param("task_id")
			task, ok := getTask(taskID)
			if !ok {
				response.Reply("Task not found")
				return
			}
			response.Reply(fmt.Sprintf("Task ID: %s\nDescription: %s\nStatus: %s",
				task.ID, task.Description, task.Status))
		},
	})

	bot.Command("update task status - <task_id> <status>", &slacker.CommandDefinition{
		Description: "Update the status of a task",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			taskID := request.Param("task_id")
			status := request.Param("status")
			task, ok := tasks[taskID]
			if !ok {
				response.Reply("Task not found")
				return
			}
			task.Status = status
			tasks[taskID] = task
			response.Reply(fmt.Sprintf("Status of task %s updated to %s", taskID, status))
		},
	})

	bot.Command("delete task - <task_id>", &slacker.CommandDefinition{
		Description: "Delete a task",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			taskID := request.Param("task_id")
			if _, ok := tasks[taskID]; !ok {
				response.Reply("Task not found")
				return
			}
			delete(tasks, taskID)
			response.Reply(fmt.Sprintf("Task %s deleted successfully", taskID))
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)

	if err != nil {
		log.Fatal(err)
	}
}

// Remaining code goes here

// Analytics
func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

// Task management functions

func createTask(description string) string {
	taskID := uuid.New().String()
	tasks[taskID] = Task{
		ID:          taskID,
		Description: description,
		Status:      "todo",
		CreatedAt:   time.Now(),
	}
	return taskID
}

func getTask(taskID string) (Task, bool) {
	task, ok := tasks[taskID]
	return task, ok
}

func getWeatherForecast(city string) (*WeatherResponse, error) {
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, openWeatherAPIKey)

	response, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var weatherData WeatherResponse
	if err := json.NewDecoder(response.Body).Decode(&weatherData); err != nil {
		return nil, err
	}

	return &weatherData, nil
}

func showTask(response slacker.ResponseWriter, taskID string) {
	task, ok := tasks[taskID]
	if !ok {
		response.Reply("Task not found")
		return
	}

	message := fmt.Sprintf("Task ID: %s\n", task.ID)
	message += fmt.Sprintf("Description: %s\n", task.Description)
	message += fmt.Sprintf("Status: %s\n", task.Status)
	message += fmt.Sprintf("Due Time: %s\n", task.DueTime.Format("2006-01-02"))
	message += fmt.Sprintf("Created At: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))

	response.Reply(message)
}

// Struct for Weather Response
type WeatherResponse struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}
