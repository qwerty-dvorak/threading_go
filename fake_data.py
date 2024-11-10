import csv
from faker import Faker

# Initialize the Faker library
fake = Faker()

# Specify the number of lines
num_lines = 1000

# Specify the CSV file name
csv_file = 'test.csv'

# Generate dummy data
dummy_data = [{'Email': fake.email(), 'Name': fake.name(), 'Age': fake.random_int(min=18, max=80), 'City': fake.city()} for _ in range(num_lines)]

# Write the dummy data to the CSV file
with open(csv_file, mode='w', newline='') as file:
    writer = csv.DictWriter(file, fieldnames=['Email', 'Name', 'Age', 'City'])
    writer.writeheader()
    writer.writerows(dummy_data)