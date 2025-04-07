import requests
from bs4 import BeautifulSoup
import pandas as pd


def scrape_active_teams():
    # URL of the page
    url = "https://www.sports-reference.com/cbb/schools/"
    
    # Add headers to mimic a browser request
    ua = ('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 '
          '(KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36')
    headers = {'User-Agent': ua}
    
    # Send a GET request to the URL
    response = requests.get(url, headers=headers)
    
    # Check if the request was successful
    if response.status_code != 200:
        msg = (f"Failed to retrieve the webpage. "
               f"Status code: {response.status_code}")
        print(msg)
        return
    
    # Parse the HTML content
    soup = BeautifulSoup(response.content, 'html.parser')
    
    # Find the table with the teams data
    table = soup.find('table', {'id': 'NCAAM_schools'})
    
    if not table:
        print("Could not find the teams table")
        return
    
    # Initialize lists to store the data
    teams_data = []
    
    # Process each row in the table
    rows = table.find_all('tr')
    
    for row in rows[1:]:  # Skip header row
        cols = row.find_all(['td', 'th'])
        if len(cols) >= 3:  # We need at least school and to columns
            school_col = row.find(['td', 'th'], {'data-stat': 'school_name'})
            to_col = row.find(['td', 'th'], {'data-stat': 'year_max'})
            
            if school_col and to_col:
                school_name = school_col.text.strip()
                to_year = to_col.text.strip()
                
                # Include teams active in 2016 or later
                try:
                    if int(to_year) >= 2016:
                        teams_data.append({
                            'School': school_name
                        })
                except ValueError:
                    # Skip if to_year isn't a valid number
                    continue
    
    # Convert to DataFrame
    df = pd.DataFrame(teams_data)
    
    # Sort by school name
    df = df.sort_values('School')
    
    # Save to CSV
    df.to_csv('active_d1_teams.csv', index=False)
    print(f"\nSuccessfully scraped {len(df)} teams active since 2016")
    return df


if __name__ == "__main__":
    scrape_active_teams() 