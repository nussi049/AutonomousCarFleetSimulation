# AutonomousCarFleetSimulation

![AutonomousCarFleetSimulation](https://github.com/nussi049/AutonomousCarFleetSimulation/assets/44850096/dab96da6-1377-456f-af15-a0d762bd2058)

This project simulates a fleet of autonomous cars navigating within a grid. Each car follows randomly generated routes, received via gRPC from a central coordinator, and updates its position in real-time on a graphical interface. The project was developed as a university assignment to demonstrate the application of gRPC, concurrency in Go, and GUI development.

## Features

- **Autonomous Car Simulation**: Multiple cars navigate within a predefined grid.
- **Random Route Generation**: Routes are generated randomly and assigned to the cars.
- **Real-time Position Updates**: Cars update their positions in real-time and can be visualized on a graphical interface.
- **gRPC Communication**: Cars receive routes and send position updates via gRPC.
- **Concurrent Processing**: The system leverages Go's concurrency model to handle multiple cars and real-time updates efficiently.
- **Graphical Interface**: A GUI built with the gioui library visualizes the grid, cars, and routes.

## Technologies Used

- **Go**: The main programming language used for the project.
- **gRPC**: For communication between the coordinator and the car clients.
- **gioui**: For creating the graphical user interface.
- **Sync & Concurrency**: Utilized Go's concurrency features for real-time updates and processing.

## Project Structure

- **coordinator**: Contains the code for the central coordinator which generates routes and sends them to the cars. Also responsible for the gui.
- **carclient**: Contains the code for the car clients which receive routes, navigate the grid, and update their positions.
- **api**: Contains the gRPC service definitions and api data structures.
- **utils**: Contains utility functions and data structures used throughout the project.

## How to Run Coordinator Client and Car Clients Manually 

1. **Setup Environment**: Ensure you have Go installed on your machine.
2. **Clone the Repository**: 
    ```sh
    git clone https://github.com/yourusername/autonomous-car-fleet-simulation.git
    cd autonomous-car-fleet-simulation
    ```
3. **Install Dependencies**: 
    ```sh
    go mod tidy
    ```
4. **Run the Coordinator**:
    ```sh
    go run coordinator/main.go
    ```
5. **Run the Car Clients**: Open multiple terminal windows and run the following command in each:
    ```sh
    go run carclient/main.go -port=<PORT> -color=<COLOR>
    ```

## Run Simulation

1. **For Quick Use**:
    ```sh
        ./start_simulation.sh <CARS> <GRIDSIZE>   
    ```
The first digit <CARS> defines the number of carclients and the second parameter <GRIDSIZE> defines the maximum grid size. 
The Grid Size should be equal to the GridSize attribute defined in the utils.go DisplaySettings struct.


## Lessons Learned

- **Concurrency in Go**: Leveraging Go's goroutines and channels for handling real-time data processing and updates.
- **gRPC Communication**: Implementing efficient communication protocols using gRPC.
- **GUI Development**: Using `gioui` for building cross-platform graphical interfaces in Go.
- **Team Collaboration**: Coordinating tasks and integrating different components of the system developed by multiple team members.

## Credits

This project was developed as part of a university assignment by:
- **Student Name 1**: Emre Sali
- **Student Name 2**: Patrick Nußbaum

---

Feel free to reach out if you have any questions or need further information about the project. Happy coding!
