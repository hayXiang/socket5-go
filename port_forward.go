package main

func port_forward(local_address *string, destination_address string) {
	process(local_address, func(_ *MyConnect) (string, error) {
		return destination_address, nil
	}, nil, nil)
}
