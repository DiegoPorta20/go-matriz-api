package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Authenticates the demo user configured in the environment and returns a bearer token.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "authentication"
                ],
                "summary": "Obtain an access token",
                "parameters": [
                    {
                        "description": "Credentials",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.LoginRequestDto"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.AccessTokenResponseDto"
                        }
                    },
                    "400": {
                        "description": "Body is not valid JSON",
                        "schema": {
                            "$ref": "#/definitions/dto.ErrorResponseDto"
                        }
                    },
                    "401": {
                        "description": "Credentials are invalid",
                        "schema": {
                            "$ref": "#/definitions/dto.ErrorResponseDto"
                        }
                    }
                }
            }
        },
        "/factorization": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Computes the QR factorization of the matrix and returns it together with the statistics of Q and R.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "factorization"
                ],
                "summary": "Factorize a matrix and calculate statistics",
                "parameters": [
                    {
                        "description": "Matrix to factorize",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.MatrixRequestDto"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.FactorizationResponseDto"
                        }
                    },
                    "400": {
                        "description": "Body is not valid JSON or matrix is missing",
                        "schema": {
                            "$ref": "#/definitions/dto.ErrorResponseDto"
                        }
                    },
                    "401": {
                        "description": "Access token is missing or invalid",
                        "schema": {
                            "$ref": "#/definitions/dto.ErrorResponseDto"
                        }
                    },
                    "422": {
                        "description": "Matrix breaks a domain rule",
                        "schema": {
                            "$ref": "#/definitions/dto.ErrorResponseDto"
                        }
                    },
                    "502": {
                        "description": "Statistics service is unavailable",
                        "schema": {
                            "$ref": "#/definitions/dto.ErrorResponseDto"
                        }
                    }
                }
            }
        },
        "/health": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Liveness probe",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.HealthResponseDto"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "dto.AccessTokenDto": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "type": "string"
                },
                "expiresIn": {
                    "type": "integer",
                    "example": 3600
                },
                "tokenType": {
                    "type": "string",
                    "example": "Bearer"
                }
            }
        },
        "dto.AccessTokenResponseDto": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/dto.AccessTokenDto"
                },
                "message": {
                    "type": "string",
                    "example": "Authentication successful"
                },
                "success": {
                    "type": "boolean",
                    "example": true
                },
                "timestamp": {
                    "type": "string",
                    "example": "2026-07-30T12:00:00Z"
                }
            }
        },
        "dto.ErrorResponseDto": {
            "type": "object",
            "properties": {
                "errors": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "message": {
                    "type": "string",
                    "example": "Invalid matrix"
                },
                "success": {
                    "type": "boolean",
                    "example": false
                },
                "timestamp": {
                    "type": "string",
                    "example": "2026-07-30T12:00:00Z"
                }
            }
        },
        "dto.FactorizationDataDto": {
            "type": "object",
            "properties": {
                "original": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "number",
                            "format": "float64"
                        }
                    }
                },
                "q": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "number",
                            "format": "float64"
                        }
                    }
                },
                "r": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "number",
                            "format": "float64"
                        }
                    }
                },
                "statistics": {
                    "$ref": "#/definitions/dto.StatisticsDto"
                }
            }
        },
        "dto.FactorizationResponseDto": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/dto.FactorizationDataDto"
                },
                "message": {
                    "type": "string",
                    "example": "Matrix processed successfully"
                },
                "success": {
                    "type": "boolean",
                    "example": true
                },
                "timestamp": {
                    "type": "string",
                    "example": "2026-07-30T12:00:00Z"
                }
            }
        },
        "dto.HealthResponseDto": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string",
                    "example": "ok"
                }
            }
        },
        "dto.LoginRequestDto": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string",
                    "example": "change-this-password"
                },
                "username": {
                    "type": "string",
                    "example": "demo"
                }
            }
        },
        "dto.MatrixRequestDto": {
            "type": "object",
            "properties": {
                "matrix": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "number",
                            "format": "float64"
                        }
                    }
                }
            }
        },
        "dto.MatrixStatisticsDto": {
            "type": "object",
            "properties": {
                "average": {
                    "type": "number",
                    "example": -0.4743
                },
                "isDiagonal": {
                    "type": "boolean",
                    "example": false
                },
                "max": {
                    "type": "number",
                    "example": 0.3162
                },
                "min": {
                    "type": "number",
                    "example": -0.9487
                },
                "sum": {
                    "type": "number",
                    "example": -1.8974
                }
            }
        },
        "dto.StatisticsDto": {
            "type": "object",
            "properties": {
                "q": {
                    "$ref": "#/definitions/dto.MatrixStatisticsDto"
                },
                "r": {
                    "$ref": "#/definitions/dto.MatrixStatisticsDto"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "description": "Bearer token obtained from /auth/login.",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Factorization API",
	Description:      "Validates a matrix, computes its QR factorization and consolidates the statistics calculated by node-api.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
