package swap

import (
	"context"
	"fmt"

	"Swaptimize/internal/worker"
)

// CreateSwapJob representa un job de creación de swap
type CreateSwapJob struct {
	id     string
	sizeMB int
	retries int
	onResult func(*JobResult)
}

// RemoveSwapJob representa un job de borrado de swap
type RemoveSwapJob struct {
	id     string
	retries int
	onResult func(*JobResult)
}

// JobResult contiene el resultado de una operación de swap
type JobResult struct {
	JobID   string
	Success bool
	Error   error
	Message string
	Started bool // Si la operación comenzó antes de error
}

// NewCreateSwapJob crea un nuevo job de creación de swap
func NewCreateSwapJob(id string, sizeMB int, onResult func(*JobResult)) worker.Job {
	return &CreateSwapJob{
		id:      id,
		sizeMB:  sizeMB,
		retries: 3,
		onResult: onResult,
	}
}

// NewRemoveSwapJob crea un nuevo job de borrado de swap
func NewRemoveSwapJob(id string, onResult func(*JobResult)) worker.Job {
	return &RemoveSwapJob{
		id:      id,
		retries: 3,
		onResult: onResult,
	}
}

// Execute ejecuta el job de creación
func (j *CreateSwapJob) Execute(ctx context.Context) error {
	result := &JobResult{
		JobID: j.id,
	}

	// Delegar a CreateSwapFileWithRetry que trata con reintentos
	err := CreateSwapFileWithRetry(ctx, j.id, j.sizeMB, j.retries)
	
	result.Success = err == nil
	result.Error = err
	if err != nil {
		result.Message = fmt.Sprintf("Failed to create swap-%s: %v", j.id, err)
	} else {
		result.Message = fmt.Sprintf("Successfully created swap-%s", j.id)
		result.Started = true
	}

	if j.onResult != nil {
		j.onResult(result)
	}

	return err
}

// Execute ejecuta el job de borrado
func (j *RemoveSwapJob) Execute(ctx context.Context) error {
	result := &JobResult{
		JobID: j.id,
	}

	err := RemoveSwapFileWithRetry(ctx, j.id, j.retries)
	
	result.Success = err == nil
	result.Error = err
	if err != nil {
		result.Message = fmt.Sprintf("Failed to remove swap-%s: %v", j.id, err)
	} else {
		result.Message = fmt.Sprintf("Successfully removed swap-%s", j.id)
		result.Started = true
	}

	if j.onResult != nil {
		j.onResult(result)
	}

	return err
}

// ID retorna el ID del job
func (j *CreateSwapJob) ID() string {
	return j.id
}

// Name retorna el nombre del job
func (j *CreateSwapJob) Name() string {
	return fmt.Sprintf("CreateSwap-%dMB", j.sizeMB)
}

// ID retorna el ID del job
func (j *RemoveSwapJob) ID() string {
	return j.id
}

// Name retorna el nombre del job
func (j *RemoveSwapJob) Name() string {
	return "RemoveSwap"
}
