/*
 * CyclopsCore.h
 *
 *  Created on: Jan 9, 2018
 *      Author: bahadir
 */

#ifndef XENONCORE_H_
#define XENONCORE_H_

#include <stdint.h>
#include <stdlib.h>

struct Results {
	const uint64_t* hashPtr;
	uint64_t size;
};


void initStorage(const uint64_t size);

void addHash(uint64_t hash);

struct Results search(const uint64_t hash, const uint64_t maxDistance);

uint64_t hashSize();

#endif /* XENONCORE_H_ */
