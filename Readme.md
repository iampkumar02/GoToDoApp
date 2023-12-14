# Todo App

This is a simple Todo application built with Go and MongoDB.

## Table of Contents

- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

### Prerequisites

Before running the Todo app, make sure you have the following installed:

- Go
- MongoDB

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/iampkumar02/GoToDoApp.git
   cd GoToDoApp
   ```

2. Install dependencies:

   ```bash
   go get
   ```

3. Start the application:

   ```bash
   go run main.go
   ```

The app should now be running at http://localhost:9000.

## Usage

- Visit http://localhost:9000 to access the Todo app.
- Use the API endpoints to interact with the Todo data.

## API Endpoints

- `GET /todo`: Fetch all Todos
- `POST /todo`: Create a new Todo
- `PUT /todo/{id}`: Update a Todo
- `DELETE /todo/{id}`: Delete a Todo
