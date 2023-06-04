# Amazon Location Service Dataset Comparison

## Overview

This small application compares an existing dataset of places or points of interest against the Amazon Location Service (ALS) dataset.

## Problem Statement

If you maintain a dataset of places or points of interest, you might want to verify if all your locations exist within the ALS dataset. This application also compares your dataset to the ALS dataset to identify any discrepancies between the two.

## Success Criteria

- **Coverage**: Every location in your dataset exists within the ALS dataset.
- **Accuracy**: The coordinates provided by ALS closely match your dataset. The closer, the better.

## Input
The application expects the following inputs:
- Place (such as a building name, street address, or Point of Interest)
- Coordinates (latitude and longitude)

## Output
The application produces the following output:
- Place (as recognized by ALS)
- Coordinates (as provided by ALS)
- The distance between your coordinates and the ones from ALS

## Example

**Input**: 

Place,Lat,Long
"Chase Plaza, Jl. Jenderal Sudirman No.21, RT.10/RW.1, Kuningan, Karet, Kecamatan Setiabudi, Kota Jakarta Selatan, Daerah Khusus Ibukota Jakarta 12920",-6.21015, 106.82187

**Output**:

Index,InputText,InputLat,InputLong,OutputText,OutputLat,OutputLong,Confidence,Distance,Status
1,"Chase Plaza, Jl. Jenderal Sudirman No.21, RT.10/RW.1, Kuningan, Karet, Kecamatan Setiabudi, Kota Jakarta Selatan, Daerah Khusus Ibukota Jakarta 12920",-6.210156,106.821871,"Chase Plaza, Jl. Jenderal Sudirman Kav. 21, Karet, Setiabudi, Kota Jakarta Selatan, Dki Jakarta, Indonesia, 12920",-6.209996,106.821800,1.000000,0.019,OK

**Insight**: ALS's data for "Chase Plaza" building is similar to the current dataset, with a difference of only 19 meters.
***
## Usage

### Prerequisites
1. Create Amazon Location Place indexes. Check the guide [here](https://docs.aws.amazon.com/location/latest/developerguide/places-prerequisites.html).
2. Provision access to the Amazon Location services via IAM Role, IAM User or SSO User.
3. Prepare an input CSV file with the following fields; Place, Lat, Long.
   - Place: The place's location, e.g., 
     - Chase Plaza, Jl. Jenderal Sudirman No.21, RT.10/RW.1, Kuningan, Karet, Kecamatan Setiabudi, Kota Jakarta Selatan, Daerah Khusus Ibukota Jakarta 12920
     - Wisma BNI 46
   - Lat: Latitude of the place
   - Long: Longitude of the place

### Steps
1. Install dependencies with `go get`.
2. Modify the `config.json` file to match your development environment:
   - ALSRegion: The region where your Amazon Location service is configured.
   - MapIndex: The index provisioned in the first step of the prerequisites.
   - Countries: Limit the search result to certain countries.
   - ConfidenceThreshold: The minimum confidence level to track.
   - InputFile: The input CSV file name from the prerequisites.
   - OutputFile: The prefix of the output file of the process.
3. Run the application using the command `go run main.go

`.
