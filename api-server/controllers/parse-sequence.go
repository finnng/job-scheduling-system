package controllers

import (
    "encoding/json"
    "errors"
    "fmt"
    "go-pg-bench/entity"
)

type ScheduleJobRequest struct {
    Steps       []map[string]interface{} `json:"steps"`
    Subscribers int                      `json:"subscribers"`
}

func ParseSequence(body ScheduleJobRequest) (*entity.Sequence, error) {
    sequence := entity.Sequence{
        Subscribers: body.Subscribers,
        Steps:       []entity.Step{},
    }
    for _, stepInterface := range body.Steps {
        step, err := UnmarshalStep(stepInterface)
        if err != nil {
            return &entity.Sequence{}, err
        }
        sequence.Steps = append(sequence.Steps, step)
    }

    return &sequence, nil
}

func UnmarshalStep(stepInterface interface{}) (entity.Step, error) {
    stepMap, ok := stepInterface.(map[string]interface{})
    if !ok {
        return nil, errors.New("invalid step format")
    }

    stepType, ok := stepMap["type"].(string)
    if !ok {
        return nil, errors.New("step type is not a string")
    }

    var step entity.Step
    var err error
    jsonData, err := json.Marshal(stepMap)
    if err != nil {
        return nil, err
    }

    switch stepType {
    case "wait_certain_period":
        step = &entity.StepWaitCertainPeriod{}
    case "wait_weekday":
        step = &entity.StepWaitWeekDay{}
    case "wait_specific_date":
        step = &entity.StepWaitSpecificDate{}
    case "job":
        step = &entity.StepJob{}
    default:
        return nil, fmt.Errorf("unsupported step type: %s", stepType)
    }

    if err = json.Unmarshal(jsonData, step); err != nil {
        return nil, err
    }

    return step, nil
}
