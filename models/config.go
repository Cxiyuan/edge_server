package models

import "strconv"

func GetConfig(key string, defaultValue string) string {
	var value string
	err := DB.QueryRow("SELECT config_value FROM system_config WHERE config_key=?", key).Scan(&value)
	if err != nil || value == "" {
		return defaultValue
	}
	return value
}

func GetConfigInt(key string, defaultValue int) int {
	value := GetConfig(key, "")
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func SetConfig(key string, value string) error {
	_, err := DB.Exec(`
		INSERT INTO system_config (config_key, config_value, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(config_key) 
		DO UPDATE SET config_value=?, updated_at=CURRENT_TIMESTAMP
	`, key, value, value)
	return err
}

func GetAllConfig() (map[string]string, error) {
	rows, err := DB.Query("SELECT config_key, config_value FROM system_config")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		config[key] = value
	}

	return config, nil
}
