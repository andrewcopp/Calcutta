# Historical Calcutta Data

This directory contains historical data from past Calcutta tournaments that have been converted from Excel files to CSV format. These files will be used to seed the database with real-world data.

## Data Conversion Process

The Excel files contain historical Calcutta data in a specific format that differs from our application's data model. The conversion process will involve:

1. **Analyzing the Excel format**: Understanding the structure and content of the existing Excel files
2. **Mapping to our data model**: Creating a mapping between the Excel data and our application's data model
3. **Converting to CSV**: Transforming the Excel data into CSV files that match our data model
4. **Data cleaning**: Ensuring data consistency and handling any anomalies
5. **Importing to database**: Creating scripts to import the CSV data into our PostgreSQL database

## Data Model Mapping

Our application uses the following data models:

- `School`: Represents a college basketball team
- `TournamentTeam`: Represents a team's participation in a specific tournament
- `Tournament`: Represents a March Madness tournament
- `User`: Represents a participant in a Calcutta
- `Calcutta`: Represents a Calcutta auction for a specific tournament
- `CalcuttaEntry`: Represents a user's entry in a Calcutta
- `CalcuttaEntryBid`: Represents a bid on a specific team
- `CalcuttaPortfolio`: Represents a user's portfolio of teams
- `CalcuttaPortfolioTeam`: Represents a team in a user's portfolio

We'll need to map the Excel data to these models, which may require:

- Splitting or combining Excel columns
- Creating relationships between entities
- Handling missing or inconsistent data
- Standardizing team names and other identifiers

## Next Steps

1. Examine the Excel files to understand their structure
2. Create a mapping document detailing how Excel data maps to our models
3. Develop conversion scripts to transform the data
4. Test the conversion process with a small sample
5. Create database import scripts
6. Validate the imported data against the original Excel files 