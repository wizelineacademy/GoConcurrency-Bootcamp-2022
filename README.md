# GoConcurrency-Bootcamp-2022 challenge by Fausto Salazar
  
## Patterns implemented

| Endpoint                   | Pattern applied |
|----------------------------|-----------------|
| `(POST) /api/provide`      | generator       |
| `(PUT) /api/refresh-cache` | fan-out/fan-in  |
 
## Execution times comparison table 

| Endpoint                   | Number of records | Exec time before solution | Exec time after solution |
|----------------------------|-------------------|---------------------------|--------------------------|
| `(POST) /api/provide`      | 100               | 25.423223773s             | 964.983809ms             |
| `(PUT) /api/refresh-cache` | 40                | 16.48707516s              | 5.998939774s             |